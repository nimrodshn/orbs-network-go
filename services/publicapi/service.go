package publicapi

import (
	"github.com/orbs-network/orbs-network-go/instrumentation"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/orbs-spec/types/go/protocol/client"
	"github.com/orbs-network/orbs-spec/types/go/services"
	"github.com/orbs-network/orbs-spec/types/go/services/handlers"
)

type service struct {
	transactionPool services.TransactionPool
	virtualMachine  services.VirtualMachine
	reporting       instrumentation.BasicLogger
}

func NewPublicApi(
	transactionPool services.TransactionPool,
	virtualMachine services.VirtualMachine,
	reporting instrumentation.BasicLogger,
) services.PublicApi {

	return &service{
		transactionPool: transactionPool,
		virtualMachine:  virtualMachine,
		reporting:       reporting.For(instrumentation.Service("public-api")),
	}
}

func (s *service) SendTransaction(input *services.SendTransactionInput) (*services.SendTransactionOutput, error) {
	s.reporting.Info("enter SendTransaction")
	defer s.reporting.Info("exit SendTransaction")
	//TODO leader should also propagate transactions to other nodes
	tx := input.ClientRequest.SignedTransaction()
	s.transactionPool.AddNewTransaction(&services.AddNewTransactionInput{
		SignedTransaction: tx,
	})

	response := &client.SendTransactionResponseBuilder{}

	return &services.SendTransactionOutput{ClientResponse: response.Build()}, nil
}

func (s *service) CallMethod(input *services.CallMethodInput) (*services.CallMethodOutput, error) {
	s.reporting.Info("enter CallMethod")
	defer s.reporting.Info("exit CallMethod")
	// TODO get block height for input ?
	rlm, err := s.virtualMachine.RunLocalMethod(&services.RunLocalMethodInput{
		Transaction: input.ClientRequest.Transaction(),
	})
	if err != nil {
		//TODO: Return graceful output on error
		return nil, nil
	}
	var oa []*protocol.MethodArgumentBuilder
	for _, arg := range rlm.OutputArguments {
		switch arg.Type() {
		case protocol.METHOD_ARGUMENT_TYPE_UINT_64_VALUE:
			oa = []*protocol.MethodArgumentBuilder{
				{Name: arg.Name(), Type: arg.Type(), Uint64Value: arg.Uint64Value()},
			}
		}
	}
	return &services.CallMethodOutput{
		ClientResponse: (&client.CallMethodResponseBuilder{
			OutputArguments: oa,
		}).Build(),
	}, nil
}

func (s *service) GetTransactionStatus(input *services.GetTransactionStatusInput) (*services.GetTransactionStatusOutput, error) {
	panic("Not implemented")
}

func (s *service) HandleTransactionResults(input *handlers.HandleTransactionResultsInput) (*handlers.HandleTransactionResultsOutput, error) {
	panic("Not implemented")
}
