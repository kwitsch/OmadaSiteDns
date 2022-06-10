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

	_ "github.com/breml/rootcerts"
	_ "github.com/kwitsch/go-dockerutils"
)

func main() {
	cfg, err := config.Get()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	l := log.New("OmadaSiteDns", (cfg.Verbose > 0))

	api, err := omadaclient.NewSiteClient(cfg.Site.Url, cfg.Site.Site,
		cfg.Site.Username, cfg.Site.Password,
		cfg.Site.SkipVerify, (cfg.Verbose > 1))

	if err != nil {
		l.E(err)
		os.Exit(2)
	}

	defer api.Close()

	cache := cache.New()

	crawler := crawler.New(api, cache, &cfg.Crawler, (cfg.Verbose > 0))
	defer crawler.Close()

	server := server.New(cache, cfg.Server, cfg.Logger, (cfg.Verbose > 0))
	defer server.Stop()
	server.Start()

	intChan := make(chan os.Signal, 1)
	signal.Notify(intChan, os.Interrupt)
	defer close(intChan)

	crawler.Start()

	for {
		select {
		case cErr := <-crawler.Error:
			l.E(cErr)
			os.Exit(3)
		case sErr := <-server.Error:
			l.E(sErr)
			os.Exit(4)
		case <-intChan:
			l.M("Server stopping")
			os.Exit(0)
		}
	}
}
