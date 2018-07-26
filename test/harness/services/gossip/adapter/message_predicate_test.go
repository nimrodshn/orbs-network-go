package adapter

import (
	"fmt"
	"github.com/orbs-network/orbs-network-go/services/gossip/adapter"
	"github.com/orbs-network/orbs-spec/types/go/primitives"
	"github.com/orbs-network/orbs-spec/types/go/protocol/consensus"
	"github.com/orbs-network/orbs-spec/types/go/protocol/gossipmessages"
)

func ExampleMessagePredicate_sender() {
	aMessageFrom := func(sender string) MessagePredicate {
		return func(data *adapter.TransportData) bool {
			return string(data.SenderPublicKey) == sender
		}
	}

	pred := aMessageFrom("sender1")

	printSender := func(sender string) {
		if pred(&adapter.TransportData{SenderPublicKey: primitives.Ed25519PublicKey(sender)}) {
			fmt.Printf("got message from %s\n", sender)
		} else {
			fmt.Println("got message from other sender")
		}
	}

	printSender("sender1")
	printSender("sender3")
	// Output: got message from sender1
	// got message from other sender
}

func ExampleMessagePredicate_payloadSize() {
	aMessageWithPayloadOver := func(maxSizeInBytes int) MessagePredicate {
		return func(data *adapter.TransportData) bool {
			size := 0
			for _, payload := range data.Payloads {
				size += len(payload)
			}

			return size < maxSizeInBytes
		}
	}

	pred := aMessageWithPayloadOver(100)

	printMessage := func(payloads [][]byte) {
		if pred(&adapter.TransportData{Payloads: payloads}) {
			fmt.Println("got message smaller than 100 bytes")
		} else {
			fmt.Println("got message larger than 100 bytes")
		}
	}

	printMessage([][]byte{make([]byte, 10)})
	printMessage([][]byte{make([]byte, 1000)})
	// Output: got message smaller than 100 bytes
	// got message larger than 100 bytes
}

func ExampleLeanHelixMessage() {
	pred := LeanHelixMessage(consensus.LEAN_HELIX_COMMIT)

	printMessage := func(msgType consensus.LeanHelixMessageType) {

		header := gossipmessages.HeaderBuilder{
			Topic:     gossipmessages.HEADER_TOPIC_LEAN_HELIX,
			LeanHelix: msgType,
		}

		if pred(&adapter.TransportData{Payloads: [][]byte{header.Build().Raw()}}) {
			fmt.Println("got commit message")
		} else {
			fmt.Println("got message of unexpected type")
		}
	}

	printMessage(consensus.LEAN_HELIX_COMMIT)
	printMessage(consensus.LEAN_HELIX_PRE_PREPARE)
	// Output: got commit message
	// got message of unexpected type
}
