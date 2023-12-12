package main

import (
	"context"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/robot/client"
	"go.viam.com/rdk/utils"
	"go.viam.com/utils/rpc"
)

func main() {

	logger := logging.NewDebugLogger("rdk-client")

	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file")
	}

	ctx := context.Background()
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	robot, err := client.New(
		ctxTimeout,
		os.Getenv("RDK_ROBOT_HOSTNAME"),
		logger,
		client.WithDialOptions(rpc.WithCredentials(rpc.Credentials{
			Type:    utils.CredentialsTypeRobotLocationSecret,
			Payload: os.Getenv("RDK_ROBOT_LOCATION_SECRET"),
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
	garagepiReturnValue, err := garagepiComponent.GPIOPinByName("37")
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Infof("garagepi GPIOPinByName return value: %+v", garagepiReturnValue)
}
