#!/usr/bin/env bash
set -euo pipefail

APP_HOME=${FOEDUS_HOME:-/app}
BIN_PATH=${FOEDUS_BIN:-/usr/local/bin/foedus}
LOG_DIR=${FOEDUS_LOG_DIR:-/var/log/foedus}
SOURCE_NODE_ID=${SOURCE_NODE_ID:-3000}
MINER_NODE_ID=${MINER_NODE_ID:-3001}
SOURCE_WAIT_SECS=${SOURCE_WAIT_SECS:-60}

SOURCE_LOG="${LOG_DIR}/source.log"
MINER_LOG="${LOG_DIR}/miner.log"
SOURCE_PID=""
MINER_PID=""
SOURCE_TAIL_PID=""
MINER_TAIL_PID=""

log() {
  printf '[BOOTSTRAP] %s\n' "$*" >&2
}

run_cli() {
  local node_id=$1
  shift
  NODE_ID="$node_id" "$BIN_PATH" "$@"
}

extract_base58_address() {
  printf '%s\n' "$1" |
    tr '\r' '\n' |
    grep -Eo '[123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]{26,}' |
    tail -n 1
}

ensure_wallet_address() {
  local node_id=$1
  local wallet_file="${APP_HOME}/temp/wallets_${node_id}.data"
  local output address

  if [[ -f "$wallet_file" ]]; then
    log "Wallet file detected for node ${node_id}, using last recorded address"
    output=$(run_cli "$node_id" listaddresses 2>&1)
    printf '%s\n' "$output" >&2
    address=$(extract_base58_address "$output")
    if [[ -z "${address:-}" ]]; then
      log "No existing addresses found; creating a new wallet"
      output=$(run_cli "$node_id" createwallet 2>&1)
      printf '%s\n' "$output" >&2
      address=$(extract_base58_address "$output")
    fi
  else
    log "Creating wallet for node ${node_id}"
    output=$(run_cli "$node_id" createwallet 2>&1)
    printf '%s\n' "$output" >&2
    address=$(extract_base58_address "$output")
  fi

  if [[ -z "${address:-}" ]]; then
    log "Failed to determine wallet address for node ${node_id}"
    exit 1
  fi

  printf '%s' "$address"
}

ensure_blockchain() {
  local node_id=$1
  local address=$2
  local db_dir="${APP_HOME}/temp/blocks_${node_id}"

  if [[ -d "$db_dir" ]] && [[ -n "$(ls -A "$db_dir" 2>/dev/null)" ]]; then
    log "Blockchain already exists for node ${node_id}, skipping genesis creation"
    return
  fi

  log "Initializing blockchain for node ${node_id}"
  run_cli "$node_id" createblockchain -address "$address"
}

start_source_server() {
  log "Starting source/API server on NODE_ID=${SOURCE_NODE_ID}"
  NODE_ID="$SOURCE_NODE_ID" "$BIN_PATH" >>"$SOURCE_LOG" 2>&1 &
  SOURCE_PID=$!
}

wait_for_source_multiaddr() {
  local elapsed=0
  while (( elapsed < SOURCE_WAIT_SECS )); do
    if [[ -f "$SOURCE_LOG" ]]; then
      local line
      line=$(grep '\[SOURCE NODE\] Address:' "$SOURCE_LOG" 2>/dev/null | head -n 1 || true)
      if [[ -n "${line:-}" ]]; then
        local addr
        addr=${line##*Address: }
        printf '%s' "$addr" | tr -d '\r' | xargs
        return 0
      fi
    fi
    sleep 1
    ((elapsed++))
  done
  return 1
}

start_miner_node() {
  local source_addr=$1
  log "Starting miner node on NODE_ID=${MINER_NODE_ID}"
  NODE_ID="$MINER_NODE_ID" "$BIN_PATH" startnode -source "$source_addr" >>"$MINER_LOG" 2>&1 &
  MINER_PID=$!
}

start_log_streams() {
  tail -n +1 -f "$SOURCE_LOG" &
  SOURCE_TAIL_PID=$!
  tail -n +1 -f "$MINER_LOG" &
  MINER_TAIL_PID=$!
}

stop_background_pids() {
  for pid in "$@"; do
    if [[ -n "${pid:-}" ]] && kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null || true
    fi
  done
}

wait_on_pids() {
  for pid in "$@"; do
    if [[ -n "${pid:-}" ]]; then
      wait "$pid" 2>/dev/null || true
    fi
  done
}

cleanup() {
  local exit_code=$1
  stop_background_pids "$SOURCE_PID" "$MINER_PID" "$SOURCE_TAIL_PID" "$MINER_TAIL_PID"
  wait_on_pids "$SOURCE_PID" "$MINER_PID" "$SOURCE_TAIL_PID" "$MINER_TAIL_PID"
  return $exit_code
}

handle_signal() {
  log "Signal received, shutting down Foedus processes"
  stop_background_pids "$SOURCE_PID" "$MINER_PID"
}

trap 'cleanup $?' EXIT
trap 'handle_signal' SIGINT SIGTERM

main() {
  mkdir -p "$APP_HOME" "$APP_HOME/temp" "$LOG_DIR"
  cd "$APP_HOME"
  : > "$SOURCE_LOG"
  : > "$MINER_LOG"

  log "Binary path: $BIN_PATH"
  log "Data directory: $APP_HOME/temp"
  log "Logs: $LOG_DIR"

  local source_wallet
  source_wallet=$(ensure_wallet_address "$SOURCE_NODE_ID")
  printf '\n' >&2
  ensure_blockchain "$SOURCE_NODE_ID" "$source_wallet"

  start_source_server
  start_log_streams

  log "Waiting for source node multiaddress (timeout: ${SOURCE_WAIT_SECS}s)"
  local multiaddr
  if ! multiaddr=$(wait_for_source_multiaddr); then
    log "Timed out waiting for source node multiaddress"
    exit 1
  fi
  log "Source multiaddress detected: $multiaddr"

  start_miner_node "$multiaddr"

  log "Source logs -> $SOURCE_LOG"
  log "Miner logs  -> $MINER_LOG"
  log "API server available on port ${SOURCE_NODE_ID}"

  if ! wait -n "$SOURCE_PID" "$MINER_PID"; then
    local status=$?
    log "One of the Foedus processes exited with status $status"
    exit $status
  fi
  log "Foedus processes exited, shutting down"
}

main
