package main

import (
	"context"
	"embed"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/cglusky/relay/pretty"
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

	prettyResourceNames := pretty.NewStringer(robot.Client.ResourceNames())
	logger.Debugf("Robot resources: %s", prettyResourceNames)

	// Create a new file system for the file server
	publicFS, err := fs.Sub(publicFiles, "public")
	if err != nil {
		logger.Fatal("Error creating public file system: ", err)
	}

	// Get the port to listen on from the environment
	// Default to 8484
	port := os.Getenv("RDK_HTTP_PORT")
	if port == "" {
		port = "8484"
	}

	// Create a new http server instance
	http.HandleFunc("/api/relay", robot.SetPinStateHandler)
	http.Handle("/", http.FileServer(http.FS(publicFS)))
	server := &http.Server{
		Addr: ":" + port,
		BaseContext: func(_ net.Listener) context.Context {
			return mainCtx
		},
	}

	// Start the http server
	go func() {
		logger.Info("Starting http server on 8484...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Error starting http server: ", err)
		}
		logger.Info("Stopped http server")
	}()

	// Block until the main context is cancelled
	logger.Info("Robot server running...")
	<-mainCtx.Done()
	logger.Info("Robot server closed")

}
