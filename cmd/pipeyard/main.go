package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Bz-Lxt/pipeyard/config"
	"github.com/Bz-Lxt/pipeyard/engine"
	"github.com/Bz-Lxt/pipeyard/httpapi"
)

func main() {
	var (
		dir = flag.String("data", "data", "yard directory")
		addr = flag.String("addr", ":8080", "listen address")
		web  = flag.String("web", "web", "static ui directory")
	)
	flag.Parse()
	cfg, err := config.Normalize(config.FromEnv(config.Config{Dir: *dir, Addr: *addr}))
	if err != nil {
		log.Fatal(err)
	}
	y, err := engine.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer y.Close()
	srv := httpapi.New(y, cfg.Addr, *web)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()
	log.Printf("pipeyard listen %s data=%s", cfg.Addr, cfg.Dir)
	if err := srv.ListenAndServe(); err != nil {
		log.Print(err)
	}
}
