package adapter

import (
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/orbs-spec/types/go/primitives"
)

type Config interface {
}

type InMemoryStatePersistence struct {
	stateWritten chan bool
	stateDiffs   map [primitives.ContractName][]*protocol.StateRecord
	config       Config
}

func NewInMemoryStatePersistence(config Config) StatePersistence {
	return &InMemoryStatePersistence{
		config:       config,
		stateDiffs:   make(map [primitives.ContractName][]*protocol.StateRecord),
		stateWritten: make(chan bool, 10),
	}
}

func (sp *InMemoryStatePersistence) WriteState(contract primitives.ContractName, stateDiff *protocol.StateRecord) {
	if contractStateDiff, ok := sp.stateDiffs[contract]; ok {
		sp.stateDiffs[contract] = append(contractStateDiff, stateDiff)
	} else {
		sp.stateDiffs[contract] = []*protocol.StateRecord{stateDiff}
	}
	sp.stateWritten <- true
}

func (sp *InMemoryStatePersistence) ReadState(contract primitives.ContractName) []*protocol.StateRecord {
	if contractStateDiff, ok := sp.stateDiffs[contract]; ok {
		return contractStateDiff
	} else {
		// TODO think about err
		return []*protocol.StateRecord{}
	}
}