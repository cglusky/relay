// package robot provides a high-level interface to the RDK client.
package robot

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/cglusky/relay/pretty"
	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/robot/client"
	"go.viam.com/rdk/utils"
	"go.viam.com/utils/rpc"
)

// Robot is a high-level interface to the RDK client and board with logger.
// Client is the RDK client.
// Board is a robot board interface.
// logger is a zap logger interface.
type Robot struct {
	Client *client.RobotClient
	Board  board.Board
	logger logging.Logger
}

// New creates a new Robot instance.
// hostname is the hostname of the robot.
// locationSecret is the location secret of the robot.
// boardName is the name of the board to use.
// Returns a Robot instance and an error.
func New(ctx context.Context, logger logging.Logger, hostname, locationSecret, boardName string) (Robot, error) {
	if hostname == "" {
		return Robot{}, errors.New("hostname must be provided")
	}

	if locationSecret == "" {
		return Robot{}, errors.New("locationSecret must be provided")
	}

	if boardName == "" {
		return Robot{}, errors.New("boardName must be provided")
	}

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
		logger: logger,
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
	logger.Debugf("Robot resources: %s", prettyResourceNames)

	return robotClient, nil
}

type pinState string

const (
	pinStateHigh pinState = "high"
	pinStateLow  pinState = "low"
)

// PinByName returns a GPIOPin by name.
// pinName is the name of the pin.
func (r Robot) PinByName(pinName string) (board.GPIOPin, error) {
	return r.Board.GPIOPinByName(pinName)
}

// GetPinState returns the state of a pin.
// pinNum is the number of the pin.
// extra is a map of extra parameters.
// Returns the state of the pin and an error.
func (r Robot) GetPinState(ctx context.Context, pinNum int, extra map[string]any) (pinState, error) {

	pinName := strconv.Itoa(pinNum)

	pin, err := r.PinByName(pinName)
	if err != nil {
		return pinStateLow, err
	}

	stateBool, err := pin.Get(ctx, extra)
	if err != nil {
		return pinStateLow, err
	}
	return boolToPinState(stateBool), nil

}

// SetPinState sets the state of a pin.
// pinNum is the number of the pin.
// state is the state to set.
// extra is a map of extra parameters.
// Returns an error.
func (r Robot) SetPinState(ctx context.Context, pinNum int, state pinState, extra map[string]any) error {
	pinName := strconv.Itoa(pinNum)

	pin, err := r.PinByName(pinName)
	if err != nil {
		return err
	}

	pinStateBool, err := pinStateToBool(state)
	if err != nil {
		return err
	}

	return pin.Set(ctx, pinStateBool, extra)
}

type RobotRequest struct {
	Action   string         `json:"action"`
	PinNum   int            `json:"pin_num"`
	PinState pinState       `json:"pin_state"`
	Extra    map[string]any `json:"extra"`
}

type RobotResponse struct {
	PinNum   int      `json:"pin_num"`
	PinState pinState `json:"pin_state"`
}

// GetPinStateHandler handles a request to get the state of a pin.
func (r Robot) GetPinStateHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	var rr RobotRequest
	err := json.NewDecoder(req.Body).Decode(&rr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pinState, err := r.GetPinState(ctx, rr.PinNum, rr.Extra)
	if err != nil {
		r.logger.Errorf("Error getting pin %d state: %s", rr.PinNum, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respBody := RobotResponse{
		PinNum:   rr.PinNum,
		PinState: pinState,
	}

	err = json.NewEncoder(w).Encode(respBody)
	if err != nil {
		r.logger.Errorf("Error encoding response body: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

}

// SetPinStateHandler handles a request to set the state of a pin.
func (r Robot) SetPinStateHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	var rr RobotRequest
	err := json.NewDecoder(req.Body).Decode(&rr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = r.SetPinState(ctx, rr.PinNum, rr.PinState, rr.Extra)
	if err != nil {
		r.logger.Errorf("Error setting pin %d state: %s", rr.PinNum, err)
	}

}

// pinStateToBool converts a pinState to a bool.
func pinStateToBool(pinState pinState) (bool, error) {
	switch pinState {
	case pinStateHigh:
		return true, nil
	case pinStateLow:
		return false, nil
	default:
		return false, errors.New("invalid pin state")
	}
}

// boolToPinState converts a bool to a pinState.
func boolToPinState(pinStateBool bool) pinState {
	if pinStateBool {
		return pinStateHigh
	}
	return pinStateLow
}

// Close closes the RDK client.
// ctx is the context.
func (r Robot) Close(ctx context.Context) {
	if r.Client == nil {
		return
	}
	r.Client.Close(ctx)
}
