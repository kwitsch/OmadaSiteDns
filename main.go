package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/kwitsch/OmadaSiteDns/cache"
	"github.com/kwitsch/OmadaSiteDns/config"
	"github.com/kwitsch/OmadaSiteDns/crawler"
	"github.com/kwitsch/OmadaSiteDns/server"
	"github.com/kwitsch/omadaclient"
	"github.com/kwitsch/omadaclient/log"

	_ "time/tzdata"

	_ "github.com/kwitsch/go-dockerutils"
)

const (
	gracefulShutdownExit = 0
	configLoadErrorExit  = 1
	omadaclientErrorExit = 2
	crawlerErrorExit     = 3
	serverErrorExit      = 4
)

func main() {
	cfg, err := config.Get()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(configLoadErrorExit)
	}

	l := log.New("OmadaSiteDns", cfg.Verbose > 0)

	api, err := omadaclient.NewSiteClient(cfg.Site.Url, cfg.Site.Site, cfg.Site.Username, cfg.Site.Password,
		cfg.Site.SkipVerify, cfg.Verbose > 1)

	if err != nil {
		l.E(err)
		os.Exit(omadaclientErrorExit)
	}

	defer api.Close()

	dnsCache := cache.New()

	crawl := crawler.New(api, dnsCache, &cfg.Crawler, cfg.Verbose > 0)
	defer crawl.Close()

	srv := server.New(dnsCache, cfg.Server, cfg.Logger, cfg.Verbose > 0)
	defer srv.Stop()
	srv.Start()

	intChan := make(chan os.Signal, 1)
	signal.Notify(intChan, os.Interrupt)
	defer close(intChan)

	crawl.Start()

	for {
		select {
		case cErr := <-crawl.Error:
			l.E(cErr)
			os.Exit(crawlerErrorExit)
		case sErr := <-srv.Error:
			l.E(sErr)
			os.Exit(serverErrorExit)
		case <-intChan:
			l.M("Server stopping")
			os.Exit(gracefulShutdownExit)
		}
	}
}
