package main

import (
	"api_gateway/internal/client"
	"api_gateway/internal/config"
	"api_gateway/internal/gateway"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	
	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load env")
	}

	cfg := config.Load()

	ctx,cancel := context.WithCancel(context.Background())
	defer cancel()

	rdb := client.NewRedis(cfg.RedisURL)

	if err := client.Ping(rdb); err != nil {
		log.Fatalf("Redis connection failed: %v",err)
	}

	log.Println("Redis connected")

	server := gateway.NewHTTPServer(ctx,cfg,rdb)

	go func() {
		log.Printf("API Gateway running on port %s",cfg.Port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v",err)
		}
	}()

	quit := make(chan os.Signal,1)
	signal.Notify(quit,syscall.SIGINT,syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(),5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("forced shutdown: %v",err)
	}

	log.Println("server exited cleanly")
}