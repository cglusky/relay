package robot

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/cglusky/relay/pretty"
	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/robot/client"
	"go.viam.com/rdk/utils"
	"go.viam.com/utils/rpc"
)

type Robot struct {
	Client *client.RobotClient
	Board  board.Board
}

func New(ctx context.Context, hostname, locationSecret, boardName string) (Robot, error) {
	if hostname == "" {
		return Robot{}, errors.New("hostname must be provided")
	}

	if locationSecret == "" {
		return Robot{}, errors.New("locationSecret must be provided")
	}

	if boardName == "" {
		return Robot{}, errors.New("boardName must be provided")
	}

	logger := logging.NewDebugLogger("rdk-client")

	robotClient, err := newClient(ctx, logger, hostname, locationSecret)
	if err != nil {
		return Robot{}, err
	}

	robotBoard, err := board.FromRobot(robotClient, boardName)
	if err != nil {
		return Robot{}, err
	}

	return Robot{
		Client: robotClient,
		Board:  robotBoard,
	}, nil
}

func newClient(ctx context.Context, logger logging.Logger, hostname string, locationSecret string) (*client.RobotClient, error) {

	ctxTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	logger.Infof("RDK client connecting to %s...", hostname)
	robotClient, err := client.New(
		ctxTimeout,
		hostname,
		logger,
		client.WithDialOptions(rpc.WithCredentials(
			rpc.Credentials{
				Type:    utils.CredentialsTypeRobotLocationSecret,
				Payload: locationSecret,
			}),
		),
	)
	if err != nil {
		return nil, err
	}
	logger.Infof("RDK client connected to %s...", hostname)

	prettyResourceNames := pretty.NewStringer(robotClient.ResourceNames())
	logger.Debugf("Resources: %s", prettyResourceNames)

	return robotClient, nil
}

func (r Robot) PinByName(pinName string) (board.GPIOPin, error) {
	return r.Board.GPIOPinByName(pinName)
}

func (r Robot) GetPinState(ctx context.Context, pinNum int, extra map[string]any) (bool, error) {

	pinName := strconv.Itoa(pinNum)

	pin, err := r.PinByName(pinName)
	if err != nil {
		return false, err
	}

	return pin.Get(ctx, extra)
}

func (r Robot) SetPinState(ctx context.Context, pinNum int, state bool, extra map[string]any) error {
	pinName := strconv.Itoa(pinNum)

	pin, err := r.PinByName(pinName)
	if err != nil {
		return err
	}

	return pin.Set(ctx, state, extra)
}

func (r Robot) Close(ctx context.Context) {
	r.Client.Close(ctx)
}
