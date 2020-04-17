package commander

import (
	"errors"
	app "twist-commander/app/interface"
	pb "twist-commander/pb"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
)

type Commander struct {
	app      app.AppImpl
	agentMgr *AgentManager
}

func CreateCommander(a app.AppImpl) *Commander {
	return &Commander{
		app:      a,
		agentMgr: CreateAgentManager(a),
	}
}

func (c *Commander) CreateRequest(transactionID string, command string, payload *any.Any) (*Agent, error) {

	agent := c.agentMgr.CreateAgent(transactionID)

	// Getting event channel
	err := agent.OpenEventChannel()
	if err != nil {
		return nil, err
	}

	// Send command
	err = agent.SendCommand(command, payload)
	if err != nil {
		agent.CloseEventChannel()
		return nil, err
	}

	return agent, nil
}

func (c *Commander) ConfirmTransaction(transactionID string, payload *pb.ConfirmTransactionRequest) error {

	data, err := ptypes.MarshalAny(payload)
	if err != nil {
		return errors.New("Failed to handle payload")
	}

	request, err := c.CreateRequest(transactionID, "confirm", data)
	if err != nil {
		return err
	}

	defer request.CloseEventChannel()

	success := false

COMPLETED:
	for {
		select {
		case event := <-request.EventChannel:

			switch event.EventName {
			case "Confirmed":
				success = true
				break COMPLETED
			case "Canceled":
				break COMPLETED
			case "Timeout":
				break COMPLETED
			}
		}
	}

	if success == false {
		return errors.New("Failed to confirm transaction")
	}

	return nil
}

func (c *Commander) RegisterTasks(transactionID string, payload *pb.RegisterTasksRequest) error {

	data, err := ptypes.MarshalAny(payload)
	if err != nil {
		return errors.New("Failed to handle payload")
	}

	request, err := c.CreateRequest(transactionID, "registerTasks", data)
	if err != nil {
		return err
	}

	defer request.CloseEventChannel()

	success := false

COMPLETED:
	for {
		select {
		case event := <-request.EventChannel:
			if event.EventName == "TasksRegistered" {
				success = true
				break COMPLETED
			}
		}
	}

	if success == false {
		return errors.New("Failed to register tasks")
	}

	return nil
}

func (c *Commander) CancelTransaction(transactionID string, payload *pb.CancelTransactionRequest) error {

	data, err := ptypes.MarshalAny(payload)
	if err != nil {
		return errors.New("Failed to handle payload")
	}

	request, err := c.CreateRequest(transactionID, "cancel", data)
	if err != nil {
		return err
	}

	defer request.CloseEventChannel()

	success := false

COMPLETED:
	for {
		select {
		case event := <-request.EventChannel:
			switch event.EventName {
			case "Canceled":
				success = true
				break COMPLETED
			case "Timeout":
				break COMPLETED
			}
		}
	}

	if success == false {
		return errors.New("Failed to register tasks")
	}

	return nil
}
