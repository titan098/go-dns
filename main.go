package main

import (
	"os"
	"os/signal"
	"syscall"

	"bitbucket.org/titan098/go-dns/config"
	"bitbucket.org/titan098/go-dns/dns"
	"bitbucket.org/titan098/go-dns/logging"
)

var log = logging.SetupLogging("main")

func cleanup(quit chan struct{}) {
	log.Info("shutting down...")
	quit <- struct{}{}
}

func main() {
	log.Info("starting up...")
	quit := make(chan struct{})

	config.Load("config.toml")

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		// trigger the quit channel so we can cleanup.
		cleanup(quit)
	}()

	// start the dns process
	dns := dns.StartServer(quit)
	defer dns.Close()

	<-quit
}
