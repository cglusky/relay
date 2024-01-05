package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/cglusky/relay/robot"
	"github.com/joho/godotenv"

	"go.viam.com/rdk/logging"
)

func main() {

	// Create a logger based on environment
	logger := logging.NewLogger("relay-main")
	if os.Getenv("RDK_PROFILE") == "development" {
		logger = logging.NewDebugLogger("relay-main")
	}

	// Create a context that is cancelled when a termination signal is received.
	// This context is parent and used to cancel all other contexts.
	mainCtx, mainCancel := context.WithCancel(context.Background())
	defer mainCancel()

	// Setup termination channel and signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	// Wait for a termination signal
	go func() {
		<-sig
		logger.Infof("Termination signal received. Stopping server...")
		mainCancel()
	}()

	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file")
	}

	robotHostname := os.Getenv("RDK_ROBOT_HOSTNAME")
	if robotHostname == "" {
		logger.Fatal("No RDK_ROBOT_HOSTNAME found in env")
	}

	robotLocationSecret := os.Getenv("RDK_ROBOT_LOCATION_SECRET")
	if robotLocationSecret == "" {
		logger.Fatal("No RDK_ROBOT_LOCATION_SECRET found in env")
	}

	robotBoardName := os.Getenv("RDK_ROBOT_BOARD_NAME")
	if robotBoardName == "" {
		logger.Fatal("No RDK_ROBOT_BOARD_NAME found in env")
	}

	// Create a new robot instance
	robot, err := robot.New(mainCtx, robotHostname, robotLocationSecret, robotBoardName)
	if err != nil {
		logger.Fatal("Error creating new robot instance: ", err)
	}
	defer robot.Close(mainCtx)

	logger.Info("Robot created")
	<-mainCtx.Done()
	logger.Info("Robot closed")

	// pinState, err := robot.GetPinState(mainCtx, 37, map[string]any{})
	// if err != nil {
	// 	logger.Errorf("Error getting pin state.  Pin:%d %s", 37, err)
	// }

	// logger.Infof("Pin State: %v", pinState)

}
