package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

const ServerInterruption = 0

func WelcomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi, I am Zijain. Thanks for assessing my interview questions!")
}

func TerminationHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "I am going to stop the server and all goroutines")
}
func StartServer(server *http.Server) error {
	return server.ListenAndServe()
}

func main() {
	defaultPort := "8081"

	// Get port from arguments
	var port string
	if len(os.Args) > 1 {
		port = os.Args[1]
	}
	if port == "" {
		port = defaultPort
	}

	// Initialize Server
	serverMux := http.NewServeMux()
	serverSigChan := make(chan int)

	// Register handler function to simulate handling normal request
	serverMux.HandleFunc("/", WelcomeHandler)
	// Register handler function to trigger server internal error
	serverMux.HandleFunc("/terminate", func(w http.ResponseWriter, r *http.Request) {
		TerminationHandler(w, r)
		serverSigChan <- ServerInterruption
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: serverMux,
	}
	// Channel for receiving shutdown signal
	sysSignalChan := make(chan os.Signal)
	signal.Notify(sysSignalChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Create error group with background context
	eg, ctx := errgroup.WithContext(context.Background())

	// Goroutine 1, which is for running the http server
	eg.Go(func() error {
		log.Println("Goroutine 1 is starting server listening on " + port)
		// Prove the server goroutine is closed
		defer log.Println("Goroutine 1 (HTTP server) has been closed")
		return StartServer(server)
	})

	// Goroutine 2 terminates the server when the errgroup is done or there is internal error inside server
	eg.Go(func() error {
		var errMessage string
		select {
		case <-ctx.Done():
			errMessage = "Errgroup is called to be terminated"
		case <-serverSigChan:
			errMessage = "Something wrong happening in server"
		}
		defer log.Println("Goroutine 2 has been closed due to: " + errMessage)
		return server.Shutdown(context.Background())
	})

	// Goroutine 3 for dealing system signal
	eg.Go(func() error {
		var errMessage string
		var err error
		select {
		case <-ctx.Done():
			errMessage = fmt.Sprintf("Something wrong happened in server")
			err = ctx.Err()
		case osSignal := <-sysSignalChan:
			errMessage = fmt.Sprintf("Received os signal(%v)", osSignal)
			err = errors.New(errMessage)
		}
		// Prove Goroutine 3 has terminated
		defer log.Println("Goroutine 3 has returned due to: " + errMessage)
		return err
	})

	err := eg.Wait()
	defer log.Println("Main process is existing due to: ", err)
}
