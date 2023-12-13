package main

import (
	"context"
	"os"
	"time"

	"github.com/cglusky/relay/pretty"
	"github.com/joho/godotenv"
	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/robot/client"
	"go.viam.com/utils/rpc"
)

func main() {

	logger := logging.NewDebugLogger("rdk-client")

	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file")
	}

	// robotHostname := os.Getenv("RDK_ROBOT_HOSTNAME")
	// if robotHostname == "" {
	// 	logger.Fatal("No RDK_ROBOT_HOSTNAME found in env")
	// }
	robotHostname := "garage.local:37931"

	ctx := context.Background()
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	logger.Infof("Client connecting to %s...", robotHostname)
	mOpts := rpc.DialMulticastDNSOptions{Disable: true}
	robot, err := client.New(
		ctxTimeout,
		robotHostname,
		logger,
		client.WithDialOptions(rpc.WithEntityCredentials(
			os.Getenv("RDK_ROBOT_API_KEY_ID"),
			rpc.Credentials{
				Type:    rpc.CredentialsTypeAPIKey,
				Payload: os.Getenv("RDK_ROBOT_API_KEY"),
			}),
			rpc.WithDialDebug(),
			rpc.WithDialMulticastDNSOptions(mOpts),
		),
	)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Client connected to %s...", robotHostname)

	defer robot.Close(ctx)

	prettyResourceNames, err := pretty.Stringer(robot.ResourceNames())
	if err != nil {
		logger.Error(err)
	} else {
		logger.Infof("Resources: %s", prettyResourceNames)
	}

	// garagepi
	rpi, err := board.FromRobot(robot, "garagepi")
	if err != nil {
		logger.Error(err)
		return
	}

	rpiGPIOPin, err := rpi.GPIOPinByName("37")
	if err != nil {
		logger.Error(err)
		return
	}

	logger.Infof("GPIOPinByName: %v", rpiGPIOPin)

	rpiGPIOPin.Set(ctx, false, map[string]interface{}{})

	time.Sleep(1 * time.Second)

	rpiGPIOPin.Set(ctx, true, map[string]interface{}{})

}
