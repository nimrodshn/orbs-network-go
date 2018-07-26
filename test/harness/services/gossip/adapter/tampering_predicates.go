package adapter

import (
	"github.com/orbs-network/orbs-network-go/services/gossip/adapter"
	"github.com/orbs-network/orbs-spec/types/go/protocol/consensus"
	"github.com/orbs-network/orbs-spec/types/go/protocol/gossipmessages"
)

// a MessagePredicate for capturing Lean Helix Consensus Algorithm gossip messages of the given type
func LeanHelixMessage(messageType consensus.LeanHelixMessageType) MessagePredicate {
	return func(data *adapter.TransportData) bool {
		header, ok := parseHeader(data)

		return ok && header.IsTopicLeanHelix() && header.LeanHelix() == messageType
	}
}

func TransactionRelayMessage(messageType gossipmessages.TransactionsRelayMessageType) MessagePredicate {
	return func(data *adapter.TransportData) bool {
		header, ok := parseHeader(data)

		return ok && header.IsTopicTransactionRelay() && header.TransactionRelay() == messageType
	}
}

func parseHeader(data *adapter.TransportData) (*gossipmessages.Header, bool) {
	if data == nil || len(data.Payloads) == 0 {
		return nil, false
	}

	header := gossipmessages.HeaderReader(data.Payloads[0])
	if !header.IsValid() {
		return nil, false
	}

	return header, true
}
