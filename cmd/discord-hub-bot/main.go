package main

import (
	"context"
	"github.com/ai-zelenin/discord-hub-bot/pkg/hub"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	logger := log.Default()

	cfg, err := hub.NewConfig()
	if err != nil {
		logger.Fatalf("Config parse err %v", err)
	}

	store, err := hub.NewFileStore(logger, cfg.DB())
	if err != nil {
		logger.Fatal(err)
	}

	b := hub.NewBot(ctx, logger, cfg, store, hub.NewSubscribeCmd(cfg, store), hub.NewUnsubscribeCmd(cfg, store))
	go func() {
		err = b.Start()
		if err != nil {
			logger.Fatalf("open err %v", err)
		}
	}()
	dispatcher := hub.NewDispatcher(logger, cfg, store, b)
	srv := hub.NewServer(cfg, logger, dispatcher)
	go func() {
		err = srv.Start()
		if err != nil {
			logger.Fatalf("server start err %v", err)
		}
	}()
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	cancel()
}
