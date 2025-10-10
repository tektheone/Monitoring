package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fleet-management-server/internal/server"
	"fleet-management-server/internal/store"
)

func main() {
	// CLI flags
	var (
		devicesFile = flag.String("devices", "./devices.csv", "Path to devices CSV file")
		bindAddr    = flag.String("bind", "127.0.0.1", "Bind address")
		port        = flag.Int("port", 6733, "Port number")
	)
	flag.Parse()

	// Initialize store and load devices from CSV
	deviceStore := store.New()
	if err := deviceStore.LoadFromCSV(*devicesFile); err != nil {
		log.Fatalf("Failed to load devices from CSV: %v", err)
	}

	log.Printf("Loaded %d devices from %s", deviceStore.DeviceCount(), *devicesFile)

	// Create HTTP server
	srv := server.New(deviceStore)
	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", *bindAddr, *port),
		Handler: srv.Router(),
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
