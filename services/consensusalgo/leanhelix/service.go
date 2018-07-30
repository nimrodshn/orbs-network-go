package leanhelix

import (
	"github.com/orbs-network/orbs-network-go/instrumentation"
	"github.com/orbs-network/orbs-spec/types/go/primitives"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/orbs-spec/types/go/protocol/consensus"
	"github.com/orbs-network/orbs-spec/types/go/services"
	"github.com/orbs-network/orbs-spec/types/go/services/gossiptopics"
	"github.com/orbs-network/orbs-spec/types/go/services/handlers"
	"sync"
	"time"
)

type Config interface {
	NetworkSize(asOfBlock uint64) uint32
	NodePublicKey() primitives.Ed25519PublicKey
	ConstantConsensusLeader() primitives.Ed25519PublicKey
	ActiveConsensusAlgo() consensus.ConsensusAlgoType
}

type service struct {
	gossip           gossiptopics.LeanHelix
	blockStorage     services.BlockStorage
	transactionPool  services.TransactionPool
	consensusContext services.ConsensusContext
	reporting        instrumentation.BasicLogger
	config           Config

	lastCommittedBlockHeight primitives.BlockHeight
	blocksForRounds          map[primitives.BlockHeight]*protocol.BlockPairContainer
	blocksForRoundsMutex     *sync.RWMutex
	votesForActiveRound      chan bool
}

func NewLeanHelixConsensusAlgo(
	gossip gossiptopics.LeanHelix,
	blockStorage services.BlockStorage,
	transactionPool services.TransactionPool,
	consensusContext services.ConsensusContext,
	reporting instrumentation.BasicLogger,
	config Config,
) services.ConsensusAlgoLeanHelix {

	s := &service{
		gossip:           gossip,
		blockStorage:     blockStorage,
		transactionPool:  transactionPool,
		consensusContext: consensusContext,
		reporting:        reporting.For(instrumentation.Service("consensus")),
		config:           config,
		lastCommittedBlockHeight: 0, // TODO: improve startup
		blocksForRounds:          make(map[primitives.BlockHeight]*protocol.BlockPairContainer),
		blocksForRoundsMutex:     &sync.RWMutex{},
	}

	gossip.RegisterLeanHelixHandler(s)
	if config.ActiveConsensusAlgo() == consensus.CONSENSUS_ALGO_TYPE_LEAN_HELIX && config.ConstantConsensusLeader().Equal(config.NodePublicKey()) {
		go s.consensusRoundRunLoop()
	}
	return s
}

func (s *service) OnNewConsensusRound(input *services.OnNewConsensusRoundInput) (*services.OnNewConsensusRoundOutput, error) {
	panic("Not implemented")
}

func (s *service) HandleTransactionsBlock(input *handlers.HandleTransactionsBlockInput) (*handlers.HandleTransactionsBlockOutput, error) {
	panic("Not implemented")
}

func (s *service) HandleResultsBlock(input *handlers.HandleResultsBlockInput) (*handlers.HandleResultsBlockOutput, error) {
	panic("Not implemented")
}

func (s *service) HandleLeanHelixPrePrepare(input *gossiptopics.LeanHelixPrePrepareInput) (*gossiptopics.EmptyOutput, error) {
	err := s.validatorVoteForNewBlockProposal(input.Message.BlockPair)
	return &gossiptopics.EmptyOutput{}, err
}

func (s *service) HandleLeanHelixPrepare(input *gossiptopics.LeanHelixPrepareInput) (*gossiptopics.EmptyOutput, error) {
	s.leaderAddVoteFromValidator()
	return &gossiptopics.EmptyOutput{}, nil
}

func (s *service) HandleLeanHelixCommit(input *gossiptopics.LeanHelixCommitInput) (*gossiptopics.EmptyOutput, error) {
	s.lastCommittedBlockHeight = s.commitBlockAndMoveToNextRound()
	return &gossiptopics.EmptyOutput{}, nil
}

func (s *service) HandleLeanHelixViewChange(input *gossiptopics.LeanHelixViewChangeInput) (*gossiptopics.EmptyOutput, error) {
	panic("Not implemented")
}

func (s *service) HandleLeanHelixNewView(input *gossiptopics.LeanHelixNewViewInput) (*gossiptopics.EmptyOutput, error) {
	panic("Not implemented")
}

func (s *service) consensusRoundRunLoop() {

	for {
		s.reporting.Info("Entered consensus round, last committed block height is", instrumentation.BlockHeight(s.lastCommittedBlockHeight))

		// see if we need to propose a new block
		err := s.leaderProposeNextBlockIfNeeded()
		if err != nil {
			s.reporting.Error(err.Error())
			continue
		}

		// validate the current proposed block
		s.blocksForRoundsMutex.RLock()
		activeBlock := s.blocksForRounds[s.lastCommittedBlockHeight+1]
		s.blocksForRoundsMutex.RUnlock()
		if activeBlock != nil {
			err := s.leaderCollectVotesForBlock(activeBlock)
			if err != nil {
				s.reporting.Error(err.Error())
				time.Sleep(10 * time.Millisecond) // TODO: handle network failures with some time of exponential backoff
				continue
			}

			// commit the block since it's validated
			s.lastCommittedBlockHeight = s.commitBlockAndMoveToNextRound()
			s.gossip.SendLeanHelixCommit(&gossiptopics.LeanHelixCommitInput{})
		}

	}
}
