package test

import (
	"context"
	"github.com/orbs-network/orbs-network-go/test"
	"github.com/orbs-network/orbs-network-go/test/builders"
	"github.com/orbs-network/orbs-network-go/test/crypto"
	"testing"
)

var privateKey = crypto.Ed25519KeyPairForTests(1).PrivateKeyUnsafe()

func TestNonLeaderDoesNotProposeBlocks(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := newHarness(false)
		h.expectNewBlockProposalNotRequested()
		h.createService(ctx)
		h.verifyNewBlockProposalNotRequested(t)
	})
}

func TestNonLeaderSavesAndRepliesToConsecutiveBlockCommits(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := newHarness(false)
		h.createService(ctx)

		aBlock := builders.BlockPair().WithBenchmarkConsensusBlockProof(privateKey, h.config.ConstantConsensusLeader())

		b1 := aBlock.WithHeight(1).Build()
		h.expectCommitSaveAndReply(b1, 1)
		h.receivedCommitViaGossip(b1)
		h.verifyCommitSaveAndReply(t)

		b2 := aBlock.WithHeight(2).WithPrevBlockHash(b1).Build()
		h.expectCommitSaveAndReply(b2, 2)
		h.receivedCommitViaGossip(b2)
		h.verifyCommitSaveAndReply(t)

		b3 := aBlock.WithHeight(3).WithPrevBlockHash(b2).Build()
		h.expectCommitSaveAndReply(b3, 3)
		h.receivedCommitViaGossip(b3)
		h.verifyCommitSaveAndReply(t)
	})
}

func TestNonLeaderSavesAndRepliesToAnOldBlockCommit(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := newHarness(false)
		h.createService(ctx)

		aBlock := builders.BlockPair().WithBenchmarkConsensusBlockProof(privateKey, h.config.ConstantConsensusLeader())

		b1 := aBlock.WithHeight(1).Build()
		h.expectCommitSaveAndReply(b1, 1)
		h.receivedCommitViaGossip(b1)
		h.verifyCommitSaveAndReply(t)

		b2 := aBlock.WithHeight(2).WithPrevBlockHash(b1).Build()
		h.expectCommitSaveAndReply(b2, 2)
		h.receivedCommitViaGossip(b2)
		h.verifyCommitSaveAndReply(t)

		// sending b1 again (an old valid block)
		h.expectCommitSaveAndReply(b1, 2)
		h.receivedCommitViaGossip(b1)
		h.verifyCommitSaveAndReply(t)
	})
}

func TestNonLeaderIgnoresFutureBlockCommit(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := newHarness(false)
		h.createService(ctx)

		aBlock := builders.BlockPair().WithBenchmarkConsensusBlockProof(privateKey, h.config.ConstantConsensusLeader())

		h.expectCommitIgnored()
		b1 := aBlock.WithHeight(1000).Build()
		h.receivedCommitViaGossip(b1)
		h.verifyCommitIgnored(t)
	})
}

func TestNonLeaderIgnoresBadPrevBlockHashPointer(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := newHarness(false)
		h.createService(ctx)

		aBlock := builders.BlockPair().WithBenchmarkConsensusBlockProof(privateKey, h.config.ConstantConsensusLeader())

		b1 := aBlock.WithHeight(1).Build()
		h.expectCommitSaveAndReply(b1, 1)
		h.receivedCommitViaGossip(b1)
		h.verifyCommitSaveAndReply(t)

		b2 := aBlock.WithHeight(2).Build()
		h.expectCommitIgnored()
		h.receivedCommitViaGossip(b2)
		h.verifyCommitIgnored(t)
	})
}

func TestNonLeaderIgnoresBadSignature(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := newHarness(false)
		h.createService(ctx)

		aBlock := builders.BlockPair().WithInvalidBenchmarkConsensusBlockProof(privateKey, h.config.ConstantConsensusLeader())

		b1 := aBlock.WithHeight(1).Build()
		h.expectCommitIgnored()
		h.receivedCommitViaGossip(b1)
		h.verifyCommitIgnored(t)
	})
}

func TestNonLeaderIgnoresBlocksFromNonLeader(t *testing.T) {
	test.WithContext(func(ctx context.Context) {
		h := newHarness(false)
		h.createService(ctx)

		aBlock := builders.BlockPair().WithBenchmarkConsensusBlockProof(privateKey, nonLeaderPublicKey())

		b1 := aBlock.WithHeight(1).Build()
		h.expectCommitIgnored()
		h.receivedCommitViaGossip(b1)
		h.verifyCommitIgnored(t)
	})
}
