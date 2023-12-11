package main

import (
	"context"

	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/robot/client"
	"go.viam.com/utils/rpc"
)

func main() {
	logger := logging.NewDebugLogger("client")
	robot, err := client.New(
		context.Background(),
		"garage-main.hq3z6kv5kx.viam.cloud",
		logger,
		client.WithDialOptions(rpc.WithEntityCredentials(
			// Replace "<API-KEY-ID>" (including brackets) with your robot's api key id
			"<API-KEY-ID>",
			rpc.Credentials{
				Type: rpc.CredentialsTypeAPIKey,
				// Replace "<API-KEY>" (including brackets) with your robot's api key
				Payload: "<API-KEY>",
			})),
	)
	if err != nil {
		logger.Fatal(err)
	}

	defer robot.Close(context.Background())
	logger.Info("Resources:")
	logger.Info(robot.ResourceNames())

	// Note that the pin supplied is a placeholder. Please change this to a valid pin.
	// garagepi
	garagepiComponent, err := board.FromRobot(robot, "garagepi")
	if err != nil {
		logger.Error(err)
		return
	}
	garagepiReturnValue, err := garagepiComponent.GPIOPinByName("16")
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Infof("garagepi GPIOPinByName return value: %+v", garagepiReturnValue)
}
