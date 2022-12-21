package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	"universe-chat/internal/agent"
	"universe-chat/internal/broker"
	"universe-chat/internal/config"
	"universe-chat/internal/nats"

	"github.com/gin-gonic/gin"
)

func main() {
	config := &config.AppConfig{}
	config.LoadConfig()
	r := gin.Default()
	natsClient, err := nats.NewClient(config.NatsConfig)
	if err != nil {
		panic(err)
	}

	broker := broker.New(natsClient)
	agent.New(r, broker)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%v:%v", config.Host, config.Port),
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}

	log.Println("Server exiting")
}
