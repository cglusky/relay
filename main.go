package main

import (
	"context"
	"os"

	"github.com/joho/godotenv"
	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/robot/client"
	"go.viam.com/utils/rpc"
)

func main() {

	logger := logging.NewDebugLogger("client")

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file")
	}

	robot, err := client.New(
		context.Background(),
		os.Getenv("RDK_ROBOT_HOSTNAME"),
		logger,
		client.WithDialOptions(rpc.WithEntityCredentials(
			os.Getenv("RDK_ROBOT_API_KEY_ID"),
			rpc.Credentials{
				Type:    rpc.CredentialsTypeAPIKey,
				Payload: os.Getenv("RDK_ROBOT_API_KEY"),
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
