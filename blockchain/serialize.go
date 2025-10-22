package blockchain

import (
	"log"

	"github.com/codila125/foedus-blockchain/protobuf"
	"google.golang.org/protobuf/proto"
)

func (outs TxOutputs) SerializeOutputs() []byte {
	/*
		Serializes the TxOutputs struct into a byte slice.
	*/
	protoOutputs := &protobuf.TxOutputs{}

	for _, output := range outs.Outputs {
		protoOutputs.Outputs = append(protoOutputs.Outputs, &protobuf.TxOutput{
			Value:      int32(output.Value),
			PubKeyHash: output.PubKeyHash,
		})
	}

	data, err := proto.Marshal(protoOutputs)
	Handle(err)
	return data
}

func DeserializeOutputs(data []byte) TxOutputs {
	/*
		Deserializes a byte slice into a TxOutputs struct.
		'data' is the byte slice to be deserialized.
	*/

	protobufOutputs := &protobuf.TxOutputs{}

	if err := proto.Unmarshal(data, protobufOutputs); err != nil {
		Handle(err)
	}

	var outputs TxOutputs
	for _, output := range protobufOutputs.Outputs {
		outputs.Outputs = append(outputs.Outputs, TxOutput{
			Value:      int(output.Value),
			PubKeyHash: output.PubKeyHash,
		})
	}
	return outputs
}

func (tx *Transaction) SerializeTransaction() []byte {
	/*
		Serializes the transaction using protobuf encoding.
		Returns the serialized byte slice.
	*/
	protoTransaction := &protobuf.Transaction{
		Id:      tx.ID,
		Inputs:  []*protobuf.TxInput{},
		Outputs: []*protobuf.TxOutput{},
	}

	for _, in := range tx.Inputs {
		protoTransaction.Inputs = append(protoTransaction.Inputs, &protobuf.TxInput{
			Id:        in.ID,
			Out:       int32(in.Out),
			Signature: in.Signature,
			PubKey:    in.PubKey,
		})
	}

	for _, out := range tx.Outputs {
		protoTransaction.Outputs = append(protoTransaction.Outputs, &protobuf.TxOutput{
			Value:      int32(out.Value),
			PubKeyHash: out.PubKeyHash,
		})
	}

	data, err := proto.Marshal(protoTransaction)
	Handle(err)

	return data
}

func DeserializeTransaction(data []byte) Transaction {
	/*
		Deserializes a byte slice into a Transaction.
	*/
	var tx protobuf.Transaction
	err := proto.Unmarshal(data, &tx)
	Handle(err)

	var inputs []TxInput
	var outputs []TxOutput

	for _, input := range tx.Inputs {
		inputs = append(inputs, TxInput{
			ID:        input.Id,
			Out:       int(input.Out),
			Signature: input.Signature,
			PubKey:    input.PubKey,
		})
	}

	for _, output := range tx.Outputs {
		outputs = append(outputs, TxOutput{
			Value:      int(output.Value),
			PubKeyHash: output.PubKeyHash,
		})
	}

	return Transaction{tx.Id, inputs, outputs}
}

func (mm *MilestoneCore) SerializeMilestoneCore() []byte {
	/*
		Serializes the immutable core of the milestone into a byte array.
	*/
	protoMilestoneCore := &protobuf.MilestoneCore{
		Title:       mm.Title,
		Description: mm.Description,
		Value:       int32(mm.Value),
		CreatedAt:   mm.CreatedAt,
	}

	data, err := proto.Marshal(protoMilestoneCore)
	Handle(err)

	return data
}

func (cs *ContractState) SerializeContractState() []byte {
	/*
		Serializes the contract state into a byte array.
	*/
	protoContractState := &protobuf.ContractState{}

	for _, ms := range cs.Milestones {
		protoMilestone := &protobuf.Milestone{
			Id:          ms.ID,
			Title:       ms.Title,
			Description: ms.Description,
			Value:       int32(ms.Value),
			DueDate:     ms.DueDate,
			Status:      string(ms.Status),
			CreatedAt:   ms.CreatedAt,
			CompletedAt: ms.CompletedAt,
			Evidence:    ms.Evidence,
			ApprovedBy:  ms.ApprovedBy,
		}
		protoContractState.Milestones = append(protoContractState.Milestones, protoMilestone)
	}

	protoContract := &protobuf.Contract{
		Id:             cs.Contract.ID,
		Title:          cs.Contract.Title,
		Description:    cs.Contract.Description,
		CreatedAt:      cs.Contract.CreatedAt,
		UpdatedAt:      cs.Contract.UpdatedAt,
		Status:         string(cs.Contract.Status),
		CreatorAddress: cs.Contract.CreatorAddress,
		Terms:          cs.Contract.Terms,
		Attachments:    cs.Contract.Attachments,
		Parties:        []*protobuf.Party{},
	}

	// Serialize Parties
	for _, p := range cs.Contract.Parties {
		protoParty := &protobuf.Party{
			Address:   p.Address,
			Role:      string(p.Role),
			PublicKey: p.PublicKey,
			Signature: p.Signature,
		}
		protoContract.Parties = append(protoContract.Parties, protoParty)
	}

	protoContractState.Contract = protoContract

	data, err := proto.Marshal(protoContractState)
	if err != nil {
		log.Panic(err)
	}
	return data
}

func DeserializeContractState(data []byte) ContractState {
	/*
		Deserializes a byte array into a ContractState.
	*/
	var state ContractState

	protoContractState := &protobuf.ContractState{}
	err := proto.Unmarshal(data, protoContractState)
	if err != nil {
		log.Panic(err)
	}

	// Deserialize Milestones
	for _, protoMs := range protoContractState.Milestones {
		ms := &Milestone{
			ID:          protoMs.Id,
			Title:       protoMs.Title,
			Description: protoMs.Description,
			Value:       int(protoMs.Value),
			DueDate:     protoMs.DueDate,
			Status:      MilestoneStatus(protoMs.Status),
			CreatedAt:   protoMs.CreatedAt,
			CompletedAt: protoMs.CompletedAt,
			Evidence:    protoMs.Evidence,
			ApprovedBy:  protoMs.ApprovedBy,
		}
		state.Milestones = append(state.Milestones, ms)
	}

	// Deserialize Contract
	protoCt := protoContractState.Contract
	contract := &Contract{
		ID:             protoCt.Id,
		Title:          protoCt.Title,
		Description:    protoCt.Description,
		CreatedAt:      protoCt.CreatedAt,
		UpdatedAt:      protoCt.UpdatedAt,
		Status:         ContractStatus(protoCt.Status),
		CreatorAddress: protoCt.CreatorAddress,
		Terms:          protoCt.Terms,
		Attachments:    protoCt.Attachments,
		Milestones:     state.Milestones,
		Parties:        []*Party{},
	}

	// Deserialize Parties from the protobuf Contract
	for _, protoP := range protoCt.Parties {
		party := &Party{
			Address:   protoP.Address,
			Role:      ContractRole(protoP.Role),
			PublicKey: protoP.PublicKey,
			Signature: protoP.Signature,
		}
		contract.Parties = append(contract.Parties, party)
	}

	state.Contract = contract

	return state
}

// SerializeContractCore serializes the immutable core of the contract into a byte array.
func (cc *ContractCore) SerializeContractCore() []byte {
	protoContractCore := &protobuf.ContractCore{
		Title:          cc.Title,
		Description:    cc.Description,
		CreatedAt:      cc.CreatedAt,
		Terms:          cc.Terms,
		CreatorAddress: cc.CreatorAddress,
		Attachments:    cc.Attachments,
	}

	for _, m := range cc.Milestones {
		protoMilestoneCore := &protobuf.MilestoneCore{
			Title:       m.Title,
			Description: m.Description,
			Value:       int32(m.Value),
			CreatedAt:   m.CreatedAt,
		}
		protoContractCore.Milestones = append(protoContractCore.Milestones, protoMilestoneCore)
	}

	for _, p := range cc.Parties {
		protoPartyCore := &protobuf.PartyCoreData{
			Address:   p.Address,
			Role:      p.Role,
			PublicKey: p.PublicKey,
		}
		protoContractCore.Parties = append(protoContractCore.Parties, protoPartyCore)
	}

	data, err := proto.Marshal(protoContractCore)
	if err != nil {
		log.Panic(err)
	}
	return data
}

func (b *Block) SerializeBlock() []byte {
	/*
		Serializes the block into a byte slice.
	*/

	protoBlock := &protobuf.Block{
		Timestamp: b.Timestamp,
		Hash:      b.Hash,
		PrevHash:  b.PrevHash,
		Nonce:     int32(b.Nonce),
		Height:    int32(b.Height),
	}

	for _, tx := range b.Transactions {
		protoTx := &protobuf.Transaction{
			Id:      tx.ID,
			Inputs:  []*protobuf.TxInput{},
			Outputs: []*protobuf.TxOutput{},
		}

		for _, in := range tx.Inputs {
			protoIn := &protobuf.TxInput{
				Id:        in.ID,
				Out:       int32(in.Out),
				Signature: in.Signature,
				PubKey:    in.PubKey,
			}
			protoTx.Inputs = append(protoTx.Inputs, protoIn)
		}

		for _, out := range tx.Outputs {
			protoOut := &protobuf.TxOutput{
				Value:      int32(out.Value),
				PubKeyHash: out.PubKeyHash,
			}
			protoTx.Outputs = append(protoTx.Outputs, protoOut)
		}

		protoBlock.Transactions = append(protoBlock.Transactions, protoTx)
	}

	for _, ct := range b.Contracts {
		protoCt := &protobuf.Contract{
			Id:             ct.ID,
			Title:          ct.Title,
			Description:    ct.Description,
			CreatorAddress: ct.CreatorAddress,
			Attachments:    ct.Attachments,
			Terms:          ct.Terms,
			Status:         string(ct.Status),
			CreatedAt:      ct.CreatedAt,
			UpdatedAt:      ct.UpdatedAt,
			Milestones:     []*protobuf.Milestone{},
			Parties:        []*protobuf.Party{},
		}

		for _, m := range ct.Milestones {
			protoMilestone := &protobuf.Milestone{
				Id:          m.ID,
				Title:       m.Title,
				Description: m.Description,
				Value:       int32(m.Value),
				DueDate:     m.DueDate,
				Status:      string(m.Status),
				CreatedAt:   m.CreatedAt,
				CompletedAt: m.CompletedAt,
				Evidence:    m.Evidence,
				ApprovedBy:  m.ApprovedBy,
			}
			protoCt.Milestones = append(protoCt.Milestones, protoMilestone)
		}

		for _, p := range ct.Parties {
			protoParty := &protobuf.Party{
				Address:   p.Address,
				Role:      string(p.Role),
				PublicKey: p.PublicKey,
				Signature: p.Signature,
			}
			protoCt.Parties = append(protoCt.Parties, protoParty)
		}

		protoBlock.Contracts = append(protoBlock.Contracts, protoCt)
	}

	data, err := proto.Marshal(protoBlock)
	Handle(err)

	return data
}

func DeserializeBlock(data []byte) *Block {
	/*
		Deserializes a byte slice into a Block.
	*/
	var block Block

	protoBlock := &protobuf.Block{}
	err := proto.Unmarshal(data, protoBlock)
	Handle(err)

	block.Timestamp = protoBlock.Timestamp
	block.Hash = protoBlock.Hash
	block.PrevHash = protoBlock.PrevHash
	block.Nonce = int(protoBlock.Nonce)
	block.Height = int(protoBlock.Height)

	for _, protoTx := range protoBlock.Transactions {
		tx := &Transaction{
			ID:      protoTx.Id,
			Inputs:  []TxInput{},
			Outputs: []TxOutput{},
		}

		for _, protoIn := range protoTx.Inputs {
			in := TxInput{
				ID:        protoIn.Id,
				Out:       int(protoIn.Out),
				Signature: protoIn.Signature,
				PubKey:    protoIn.PubKey,
			}
			tx.Inputs = append(tx.Inputs, in)
		}

		for _, protoOut := range protoTx.Outputs {
			out := TxOutput{
				Value:      int(protoOut.Value),
				PubKeyHash: protoOut.PubKeyHash,
			}
			tx.Outputs = append(tx.Outputs, out)
		}

		block.Transactions = append(block.Transactions, tx)
	}

	for _, protoCt := range protoBlock.Contracts {
		ct := &Contract{
			ID:             protoCt.Id,
			Title:          protoCt.Title,
			Description:    protoCt.Description,
			CreatorAddress: protoCt.CreatorAddress,
			Attachments:    protoCt.Attachments,
			Terms:          protoCt.Terms,
			Status:         ContractStatus(protoCt.Status),
			CreatedAt:      protoCt.CreatedAt,
			UpdatedAt:      protoCt.UpdatedAt,
			Milestones:     []*Milestone{},
			Parties:        []*Party{},
		}

		for _, protoM := range protoCt.Milestones {
			m := &Milestone{
				ID:          protoM.Id,
				Title:       protoM.Title,
				Description: protoM.Description,
				Value:       int(protoM.Value),
				DueDate:     protoM.DueDate,
				Status:      MilestoneStatus(protoM.Status),
				CreatedAt:   protoM.CreatedAt,
				CompletedAt: protoM.CompletedAt,
				Evidence:    protoM.Evidence,
				ApprovedBy:  protoM.ApprovedBy,
			}
			ct.Milestones = append(ct.Milestones, m)
		}

		for _, protoP := range protoCt.Parties {
			p := &Party{
				Address:   protoP.Address,
				Role:      ContractRole(protoP.Role),
				PublicKey: protoP.PublicKey,
				Signature: protoP.Signature,
			}
			ct.Parties = append(ct.Parties, p)
		}

		block.Contracts = append(block.Contracts, ct)
	}

	return &block
}
