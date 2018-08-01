package adapter

import (
	"github.com/orbs-network/orbs-spec/types/go/primitives"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/pkg/errors"
)

type Config interface {
}

type InMemoryStatePersistence struct {
	stateWritten chan bool
	stateDiffs   map[primitives.ContractName]map[string]*protocol.StateRecord
	config       Config
}

func NewInMemoryStatePersistence(config Config) StatePersistence {
	return &InMemoryStatePersistence{
		config: config,
		// TODO remove init with a hard coded contract once deploy/provisioning of contracts exists
		stateDiffs:   map[primitives.ContractName]map[string]*protocol.StateRecord{primitives.ContractName("BenchmarkToken"): {}},
		stateWritten: make(chan bool, 10),
	}
}

func (sp *InMemoryStatePersistence) WriteState(contract primitives.ContractName, stateDiff *protocol.StateRecord) error {
	if _, ok := sp.stateDiffs[contract]; !ok {
		sp.stateDiffs[contract] = map[string]*protocol.StateRecord{}
	}

	sp.stateDiffs[contract][stateDiff.Key().KeyForMap()] = stateDiff
	sp.stateWritten <- true

	return nil
}

func (sp *InMemoryStatePersistence) ReadState(contract primitives.ContractName) (map[string]*protocol.StateRecord, error){
	if contractStateDiff, ok := sp.stateDiffs[contract]; ok {
		return contractStateDiff, nil
	} else {
		return nil, errors.Errorf("contract %v does not exist", contract)
	}
}
