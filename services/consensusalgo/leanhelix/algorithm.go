package leanhelix

import (
	"fmt"
	"github.com/orbs-network/orbs-network-go/services/blockstorage"
	"github.com/orbs-network/orbs-spec/types/go/primitives"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/orbs-spec/types/go/protocol/consensus"
	"github.com/orbs-network/orbs-spec/types/go/protocol/gossipmessages"
	"github.com/orbs-network/orbs-spec/types/go/services"
	"github.com/orbs-network/orbs-spec/types/go/services/gossiptopics"
)

func (s *service) leaderProposeNextBlockIfNeeded() error {
	nextBlockHeight := s.lastCommittedBlockHeight + 1

	s.blocksForRoundsMutex.RLock()
	nextBlock := s.blocksForRounds[nextBlockHeight]
	s.blocksForRoundsMutex.RUnlock()
	if nextBlock != nil {
		return nil
	}

	proposedTransactions, err := s.transactionPool.GetTransactionsForOrdering(&services.GetTransactionsForOrderingInput{
		MaxNumberOfTransactions: 1,
	})
	if err != nil {
		return err
	}

	proposedBlockPair := &protocol.BlockPairContainer{
		TransactionsBlock: &protocol.TransactionsBlockContainer{
			Header: (&protocol.TransactionsBlockHeaderBuilder{
				ProtocolVersion:       blockstorage.ProtocolVersion,
				BlockHeight:           primitives.BlockHeight(s.lastCommittedBlockHeight + 1),
				NumSignedTransactions: uint32(len(proposedTransactions.SignedTransactions)),
			}).Build(),
			Metadata:           (&protocol.TransactionsBlockMetadataBuilder{}).Build(),
			SignedTransactions: proposedTransactions.SignedTransactions,
			BlockProof:         (&protocol.TransactionsBlockProofBuilder{}).Build(),
		},
		ResultsBlock: &protocol.ResultsBlockContainer{
			Header: (&protocol.ResultsBlockHeaderBuilder{
				ProtocolVersion: blockstorage.ProtocolVersion,
				BlockHeight:     primitives.BlockHeight(s.lastCommittedBlockHeight + 1),
			}).Build(),
			TransactionReceipts: nil,
			ContractStateDiffs:  nil,
			BlockProof:          (&protocol.ResultsBlockProofBuilder{}).Build(),
		},
	}

	s.blocksForRoundsMutex.Lock()
	s.blocksForRounds[nextBlockHeight] = proposedBlockPair
	s.blocksForRoundsMutex.Unlock()

	s.reporting.Infof("Proposed block pair for height %d", nextBlockHeight)

	return nil
}

func (s *service) leaderCollectVotesForBlock(blockPair *protocol.BlockPairContainer) error {
	s.votesForActiveRound = make(chan bool)
	defer func() {
		close(s.votesForActiveRound)
		s.votesForActiveRound = nil
	}()

	_, err := s.gossip.SendLeanHelixPrePrepare(&gossiptopics.LeanHelixPrePrepareInput{
		Message: &gossipmessages.LeanHelixPrePrepareMessage{
			SignedHeader: (&consensus.LeanHelixBlockRefBuilder{}).Build(),
			Sender:       (&consensus.LeanHelixSenderSignatureBuilder{}).Build(),
			BlockPair:    blockPair,
		},
	})
	if err != nil {
		return err
	}

	// asking for votes from everybody except ourselves
	numOfRequiredVotes := int(s.config.NetworkSize(0)) - 1
	for i := 0; i < numOfRequiredVotes; i++ {
		<-s.votesForActiveRound
	}

	s.reporting.Infof("Got the required %d votes for next block", numOfRequiredVotes)

	return nil
}

func (s *service) validatorVoteForNewBlockProposal(blockPair *protocol.BlockPairContainer) error {
	blockHeight := blockPair.TransactionsBlock.Header.BlockHeight()

	s.blocksForRoundsMutex.Lock()
	s.blocksForRounds[blockHeight] = blockPair
	s.blocksForRoundsMutex.Unlock()

	s.reporting.Infof("Voting as validator for block of height %d", blockHeight)
	_, err := s.gossip.SendLeanHelixPrepare(&gossiptopics.LeanHelixPrepareInput{})
	return err
}

func (s *service) leaderAddVoteFromValidator() {
	// TODO: we assume we only get votes for the active round, in the real world we can't assume this,
	// TODO:  but here since we don't move to the next round unless everybody voted, it's ok
	if s.votesForActiveRound == nil {
		panic("received vote while not collecting votes")
	}
	s.votesForActiveRound <- true
}

func (s *service) commitBlockAndMoveToNextRound() primitives.BlockHeight {
	blockHeight := s.lastCommittedBlockHeight + 1

	s.blocksForRoundsMutex.RLock()
	blockPair, found := s.blocksForRounds[blockHeight]
	s.blocksForRoundsMutex.RUnlock()

	if !found {
		err := fmt.Errorf("trying to commit a block of height %d that wasn't prepared", blockHeight)
		s.reporting.Error(err)
		panic(err)
	}

	s.blockStorage.CommitBlock(&services.CommitBlockInput{
		BlockPair: blockPair,
	})

	s.blocksForRoundsMutex.Lock()
	delete(s.blocksForRounds, blockHeight)
	s.blocksForRoundsMutex.Unlock()

	return blockHeight
}
