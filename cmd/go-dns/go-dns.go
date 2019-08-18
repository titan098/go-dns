package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/titan098/go-dns/config"
	"github.com/titan098/go-dns/dns"
	"github.com/titan098/go-dns/logging"
)

var log = logging.SetupLogging("main")

func cleanup(quit chan struct{}) {
	log.Info("shutting down...")
	quit <- struct{}{}
}

func main() {
	log.Info("starting up...")

	// define the input flags
	var configFile string
	flag.StringVar(&configFile, "c", "", "the paht to the config file.")
	flag.Parse()

	// find out if are in a snap and we can load the config from
	// the data directory
	if configFile == "" {
		configFile = config.LocateConfigFile()
	}

	config.Load(configFile)

	quit := make(chan struct{})
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
