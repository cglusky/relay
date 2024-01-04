package robot

import (
	"context"
	"errors"
	"time"

	"github.com/cglusky/relay/pretty"
	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/robot/client"
	"go.viam.com/rdk/utils"
	"go.viam.com/utils/rpc"
)

type Robot struct {
	hostname       string
	locationSecret string
	logger         logging.Logger
	client         *client.RobotClient
}

func New(hostname, locationSecret string) (Robot, error) {
	if hostname == "" || locationSecret == "" {
		return Robot{}, errors.New("hostname and locationSecret must be provided")
	}

	logger := logging.NewDebugLogger("rdk-client")
	ctx := context.Background()

	robotClient, err := newClient(ctx, logger, hostname, locationSecret)
	if err != nil {
		return Robot{}, err
	}

	return Robot{
		hostname:       hostname,
		locationSecret: locationSecret,
		logger:         logger,
		client:         robotClient,
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
		logger.Error("err")
	}
	logger.Infof("RDK client connected to %s...", hostname)

	prettyResourceNames := pretty.NewStringer(robotClient.ResourceNames())
	logger.Debugf("Resources: %s", prettyResourceNames)
	return robotClient, nil

}

func (r Robot) boardByName(name string) (board.Board, error) {
	return board.FromRobot(r.client, name)
}

func (r Robot) PinByName(boardName, pinName string) (board.GPIOPin, error) {
	board, err := r.boardByName(boardName)
	if err != nil {
		return nil, err
	}
	return board.GPIOPinByName(pinName)
}

// rpiGPIOPin, err := rpi.GPIOPinByName("37")
// if err != nil {
// 	logger.Error(err)
// 	return
// }

// logger.Infof("GPIOPinByName: %v", rpiGPIOPin)

// rpiGPIOPin.Set(ctx, false, map[string]interface{}{})

// time.Sleep(1 * time.Second)

// rpiGPIOPin.Set(ctx, true, map[string]interface{}{})

func (r *Robot) Close(ctx context.Context) {
	r.client.Close(ctx)
}
