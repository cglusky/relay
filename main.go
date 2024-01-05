package main

import (
	"context"
	"embed"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/cglusky/relay/robot"
	"github.com/joho/godotenv"

	"go.viam.com/rdk/logging"
)

//go:embed public
var publicFiles embed.FS

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
	robot, err := robot.New(
		mainCtx,
		logger,
		robotHostname,
		robotLocationSecret,
		robotBoardName,
	)
	if err != nil {
		logger.Fatal("Error creating new robot instance: ", err)
	}
	defer robot.Close(mainCtx)

	// Create a new http server instance
	http.HandleFunc("/api/relay", robot.SetPinStateHandler)
	http.Handle("/", http.FileServer(http.FS(publicFiles)))
	server := &http.Server{
		Addr: ":8080",
		BaseContext: func(_ net.Listener) context.Context {
			return mainCtx
		},
	}

	// Start the http server
	go func() {
		logger.Info("Starting http server on 8080...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Error starting http server: ", err)
		}
		logger.Info("Stopped http server")
	}()

	logger.Info("Robot server running...")
	<-mainCtx.Done()
	logger.Info("Robot server closed")

}
