package crawler

import (
	"regexp"
	"strings"
	"time"

	"github.com/kwitsch/OmadaSiteDns/cache"
	"github.com/kwitsch/OmadaSiteDns/config"
	"github.com/kwitsch/OmadaSiteDns/osdutils"
	"github.com/kwitsch/omadaclient/apiclient"
	"github.com/kwitsch/omadaclient/log"
)

type Crawler struct {
	api   *apiclient.Apiclient
	cache *cache.Cache
	cfg   *config.Crawler
	l     *log.Log
	Error chan (error)
}

func New(api *apiclient.Apiclient, cache *cache.Cache, cfg *config.Crawler, verbose bool) *Crawler {
	return &Crawler{
		api:   api,
		cache: cache,
		cfg:   cfg,
		l:     log.New("Crawler", verbose),
		Error: make(chan error, 2),
	}
}

func (c *Crawler) Start() {
	go func() {
		c.fetch()

		ticker := time.NewTicker(c.cfg.Intervall)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.fetch()
			}
		}
	}()
}

func (c *Crawler) Close() {
	close(c.Error)
}

func (c *Crawler) fetch() {
	c.l.M("Start fetching data", time.Now())

	clients, err := c.api.Clients()
	if c.failed(err) {
		return
	}

	defer c.api.Close()

	c.l.V("Fetched Clients:", len(*clients))

	for _, cl := range *clients {
		hname := c.convertName(cl.Name)
		c.l.V("Fetch:", cl.IP, "-", hname)
		if osdutils.ValidDnsStr(hname) {
			c.cache.Update(hname, cl.IP)
		} else {
			c.l.M("Invalid hostname:", hname)
		}
	}

	if c.cfg.Gateway.Include {
		c.l.V("Fetch gateway")
		devices, err := c.api.DevicesDetailed()
		if c.failed(err) {
			return
		}

		ips := []string{}
		for _, d := range *devices {
			if d.Type == "gateway" {
				gname := c.convertName(d.Name)
				c.l.V("Gateway name:", gname)
				for _, lcs := range d.LanClientStats {
					c.l.V("-", lcs.LanName, ":", lcs.IP)
					if len(c.cfg.Gateway.PrimaryNet) > 0 && c.cfg.Gateway.PrimaryNet == lcs.LanName {
						ips = append([]string{lcs.IP}, ips...)
					} else {
						ips = append(ips, lcs.IP)
					}
				}

				c.l.V("Primary address:", ips[0])

				c.cache.Update(gname, ips...)
				break
			}
		}
	}

	c.l.ReturnSuccess()
}

func (c *Crawler) failed(err error) bool {
	if err != nil {
		c.Error <- err
		return true
	}
	return false
}

func (c *Crawler) convertName(name string) string {
	res := strings.ToLower(name)

	for _, conv := range c.cfg.Converters {
		cconf, err := regexp.Compile(conv.Regex)
		if c.failed(err) {
			break
		}
		res = cconf.ReplaceAllString(res, conv.Substitute)
	}

	return osdutils.RemoveInvalidCharacters(res)
}
