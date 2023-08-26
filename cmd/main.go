package main

import (
	"flag"
	"main/config"
	"main/internal/fetcher"
	"main/internal/handler"
	"main/internal/storage"
	"main/pkg/postgres"
	"main/pkg/telegram"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

func main() {
	file := flag.String("config", "./config/config.yaml", "config file name path")
	flag.Parse()

	ctx := context.Background()

	c, err := config.NewConfig(*file)
	if err != nil {
		log.Fatalf("cannot create new config: %v", err)
	}
	b, err := telegram.NewBot(ctx, c.Telegram)
	if err != nil {
		log.Fatalf("cannot create new telegram bot: %v", err)
	}
	p, err := postgres.NewClient(ctx, c.Postgres)
	if err != nil {
		log.Fatalf("cannot create new postgres client: %v", err)
	}
	s := storage.NewStorage(ctx, p)
	f := fetcher.NewFetcher(ctx)

	h, err := handler.NewHandler(ctx, b, f, s)
	if err != nil {
		log.Fatalf("cannot create new handler: %v", err)
	}
	defer h.Shutdown(ctx)

	go h.HandleMessagesContinuously(ctx)
	go h.HandleSubscriptionsContinuously(ctx)
	go h.HandleTasksContinuously(ctx)

	time.Sleep(10 * time.Second)
	h.Shutdown(ctx)

	exit := make(chan os.Signal)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)
	<-exit
}
