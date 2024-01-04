// package robot provides a high-level interface to the RDK client.
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

// Robot is a high-level interface to the RDK client.
// Client is the RDK client.
// Board is the robot board.
type Robot struct {
	Client *client.RobotClient
	Board  board.Board
}

// New creates a new Robot instance.
// hostname is the hostname of the robot.
// locationSecret is the location secret of the robot.
// boardName is the name of the board to use.
// Returns a Robot instance and an error.
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

// newClient creates a new RDK client.
// ctx is the context.
// hostname is the hostname of the robot.
// locationSecret is the location secret of the robot.
// Returns a RobotClient instance and an error.
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

// PinByName returns a GPIOPin by name.
// pinName is the name of the pin.
func (r Robot) PinByName(pinName string) (board.GPIOPin, error) {
	return r.Board.GPIOPinByName(pinName)
}

// GetPinState returns the state of a pin.
// pinNum is the number of the pin.
// extra is a map of extra parameters.
// Returns the state of the pin and an error.
func (r Robot) GetPinState(ctx context.Context, pinNum int, extra map[string]any) (bool, error) {

	pinName := strconv.Itoa(pinNum)

	pin, err := r.PinByName(pinName)
	if err != nil {
		return false, err
	}

	return pin.Get(ctx, extra)
}

// SetPinState sets the state of a pin.
// pinNum is the number of the pin.
// state is the state to set.
// extra is a map of extra parameters.
// Returns an error.
func (r Robot) SetPinState(ctx context.Context, pinNum int, state bool, extra map[string]any) error {
	pinName := strconv.Itoa(pinNum)

	pin, err := r.PinByName(pinName)
	if err != nil {
		return err
	}

	return pin.Set(ctx, state, extra)
}

// Close closes the RDK client.
// ctx is the context.
func (r Robot) Close(ctx context.Context) {
	r.Client.Close(ctx)
}
