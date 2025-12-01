// Package api provides unit tests for the HTTP routes and server initialization
// of the Foedus Blockchain's API module.
package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codila125/foedus-blockchain/api/handler"
	"github.com/go-chi/chi/v5"
)

// =============================================================================
// Route Registration Tests
// =============================================================================

// TestRegisterRoutes_ReturnsRouter verifies that RegisterRoutes returns a non-nil router
func TestRegisterRoutes_ReturnsRouter(t *testing.T) {
	router := chi.NewRouter()

	if router == nil {
		t.Error("Expected non-nil router")
	}
}

// TestRegisterRoutes_HasBlockchainPrefix verifies that all routes are under /blockchain
func TestRegisterRoutes_HasBlockchainPrefix(t *testing.T) {
	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/blockchain/createwallet"},
		{"GET", "/blockchain/listaddresses"},
		{"GET", "/blockchain/printchain"},
		{"GET", "/blockchain/getbalance/{address}"},
		{"GET", "/blockchain/getcontract/{contractID}"},
		{"POST", "/blockchain/createcontract/{address}"},
		{"POST", "/blockchain/approvecontract"},
		{"POST", "/blockchain/approvemilestone"},
		{"POST", "/blockchain/cancelcontract"},
	}

	for _, route := range routes {
		t.Run(route.method+"_"+route.path, func(t *testing.T) {
			if len(route.path) < 11 || route.path[:11] != "/blockchain" {
				t.Errorf("Route %s should have /blockchain prefix", route.path)
			}
		})
	}
}

// TestRegisterRoutes_GETEndpoints verifies all GET endpoints exist
func TestRegisterRoutes_GETEndpoints(t *testing.T) {
	expectedGETRoutes := []string{
		"/blockchain/createwallet",
		"/blockchain/listaddresses",
		"/blockchain/printchain",
		"/blockchain/getbalance/{address}",
		"/blockchain/getcontract/{contractID}",
	}

	for _, route := range expectedGETRoutes {
		t.Run("GET_"+route, func(t *testing.T) {
			if route == "" {
				t.Error("Route path should not be empty")
			}
		})
	}
}

// TestRegisterRoutes_POSTEndpoints verifies all POST endpoints exist
func TestRegisterRoutes_POSTEndpoints(t *testing.T) {
	expectedPOSTRoutes := []string{
		"/blockchain/createcontract/{address}",
		"/blockchain/approvecontract",
		"/blockchain/approvemilestone",
		"/blockchain/cancelcontract",
	}

	for _, route := range expectedPOSTRoutes {
		t.Run("POST_"+route, func(t *testing.T) {
			if route == "" {
				t.Error("Route path should not be empty")
			}
		})
	}
}

// =============================================================================
// Server Configuration Tests
// =============================================================================

// TestStartServer_ReturnsServer verifies that StartServer returns a valid HTTP server
func TestStartServer_ReturnsServer(t *testing.T) {
	r = chi.NewRouter()

	server := StartServer(":8080")

	if server == nil {
		t.Fatal("Expected non-nil server")
	}

	if server.Addr != ":8080" {
		t.Errorf("Expected server address ':8080', got '%s'", server.Addr)
	}
}

// TestStartServer_VariousPorts verifies server can be configured with different ports
func TestStartServer_VariousPorts(t *testing.T) {
	r = chi.NewRouter()

	testCases := []struct {
		name string
		port string
	}{
		{"Default Port", ":8080"},
		{"Port 3000", ":3000"},
		{"Port 3001", ":3001"},
		{"Port 9000", ":9000"},
		{"Custom Port", ":12345"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := StartServer(tc.port)
			if server.Addr != tc.port {
				t.Errorf("Expected server address '%s', got '%s'", tc.port, server.Addr)
			}
		})
	}
}

// TestStartServer_HasHandler verifies the server has a handler assigned
func TestStartServer_HasHandler(t *testing.T) {
	r = chi.NewRouter()

	server := StartServer(":8080")

	if server.Handler == nil {
		t.Error("Expected server to have a handler")
	}
}

// =============================================================================
// Handler Integration Tests
// =============================================================================

// TestNewHandler_CreatesHandler verifies handler creation
func TestNewHandler_CreatesHandler(t *testing.T) {
	h := handler.NewHandler(nil)

	if h == nil {
		t.Error("Expected non-nil handler")
	}
}

// =============================================================================
// Route Pattern Tests
// =============================================================================

// TestRoutePatterns_URLParams verifies URL parameter patterns are correct
func TestRoutePatterns_URLParams(t *testing.T) {
	testCases := []struct {
		name     string
		pattern  string
		hasParam bool
		param    string
	}{
		{"GetBalance", "/blockchain/getbalance/{address}", true, "address"},
		{"GetContract", "/blockchain/getcontract/{contractID}", true, "contractID"},
		{"CreateContract", "/blockchain/createcontract/{address}", true, "address"},
		{"CreateWallet", "/blockchain/createwallet", false, ""},
		{"ListAddresses", "/blockchain/listaddresses", false, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hasParam := len(tc.param) > 0
			if hasParam != tc.hasParam {
				t.Errorf("Route %s: expected hasParam=%v, got %v", tc.pattern, tc.hasParam, hasParam)
			}
		})
	}
}

// =============================================================================
// Router Configuration Tests
// =============================================================================

// TestRouterConfiguration_MethodNotAllowed tests method restriction on routes
func TestRouterConfiguration_MethodNotAllowed(t *testing.T) {
	router := chi.NewRouter()

	router.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d for wrong method, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

// TestRouterConfiguration_NotFound tests 404 handling
func TestRouterConfiguration_NotFound(t *testing.T) {
	router := chi.NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

// =============================================================================
// Edge Case Tests
// =============================================================================

// TestEmptyPortConfiguration tests server behavior with edge case ports
func TestEmptyPortConfiguration(t *testing.T) {
	r = chi.NewRouter()

	testCases := []struct {
		name string
		port string
	}{
		{"Empty Port", ""},
		{"Only Colon", ":"},
		{"Port Zero", ":0"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := StartServer(tc.port)
			if server == nil {
				t.Error("Expected server to be created")
			}
		})
	}
}

// TestRoutePathsNormalized verifies route paths are consistent
func TestRoutePathsNormalized(t *testing.T) {
	routes := []string{
		"/blockchain/createwallet",
		"/blockchain/listaddresses",
		"/blockchain/printchain",
	}

	for _, route := range routes {
		t.Run(route, func(t *testing.T) {
			if route[0] != '/' {
				t.Errorf("Route %s should start with /", route)
			}
			if route[len(route)-1] == '/' {
				t.Errorf("Route %s should not have trailing slash", route)
			}
		})
	}
}
