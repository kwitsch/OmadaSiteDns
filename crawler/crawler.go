package crawler

import (
	"regexp"
	"strings"
	"time"

	"github.com/kwitsch/OmadaSiteDns/cache"
	"github.com/kwitsch/OmadaSiteDns/config"
	"github.com/kwitsch/OmadaSiteDns/domainstore"
	"github.com/kwitsch/OmadaSiteDns/osdutils"
	"github.com/kwitsch/omadaclient"
	"github.com/kwitsch/omadaclient/log"
)

type Crawler struct {
	api     *omadaclient.SiteClient
	cache   *cache.Cache
	cfg     *config.Crawler
	l       *log.Log
	domains *domainstore.Domainstore
	Error   chan (error)
}

func New(api *omadaclient.SiteClient, cache *cache.Cache, cfg *config.Crawler, verbose bool) *Crawler {
	res := Crawler{
		api:     api,
		cache:   cache,
		cfg:     cfg,
		l:       log.New("Crawler", verbose),
		domains: domainstore.New(verbose),
		Error:   make(chan error, 2),
	}

	for _, n := range cfg.Network {
		res.domains.AddOverride(n.Name, n.Domain)
	}

	return &res
}

func (c *Crawler) Start() {
	go func() {
		c.fetchNetworks()
		c.fetchHosts()

		hostTicker := time.NewTicker(c.cfg.HostIntervall)
		defer hostTicker.Stop()

		networkTicker := time.NewTicker(c.cfg.NetworkIntervall)
		defer networkTicker.Stop()

		for {
			select {
			case <-networkTicker.C:
				hostTicker.Reset(c.cfg.HostIntervall)
				c.fetchNetworks()
			case <-hostTicker.C:
				c.fetchHosts()
			}
		}
	}()
}

func (c *Crawler) Close() {
	close(c.Error)
}

func (c *Crawler) fetchNetworks() {
	c.l.M("Start fetching networks", time.Now())

	networks, err := c.api.GetNetworks()
	if c.failed(err) {
		return
	}

	defer c.api.Close()

	for _, n := range *networks {
		err := c.domains.AddNetwork(n.Name, n.GatewaySubnet, n.Domain)
		c.l.V(n.Name, err)
	}
}

func (c *Crawler) fetchHosts() {
	c.l.M("Start fetching hosts", time.Now())

	clients, err := c.api.GetClients(false)
	if c.failed(err) {
		return
	}

	defer c.api.Close()

	c.l.V("Fetched Clients:", len(*clients))

	for _, cl := range *clients {
		domname, ok := c.getDomName(cl.Name, cl.IP)
		if ok {
			c.addIfValid(domname, cl.IP)
		} else {
			c.l.V("No network found for IP:", cl.IP)
		}
	}

	c.l.V("Fetch devices")
	devices, err := c.api.GetDevices(true)
	if c.failed(err) {
		return
	}

	for _, d := range *devices {
		if d.Type == "gateway" {
			for _, lcs := range d.LanClientStats {
				domname, ok := c.getDomName(d.Name, lcs.IP)
				if ok {
					c.addIfValid(domname, lcs.IP)
				} else {
					c.l.V("No network found for IP:", lcs.IP)
				}
			}

			break
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

func (c *Crawler) getDomName(name, ip string) (string, bool) {
	dom, ok := c.domains.GetDomain(ip)
	if ok {
		res := c.convertName(name)
		if len(dom) > 0 {
			res += "." + dom
		}
		return res, true
	}
	return "", false
}

func (c *Crawler) addIfValid(name, ip string) {
	c.l.V("Fetch:", ip, "-", name)
	if osdutils.ValidDnsStr(name) {
		c.cache.Update(name, ip)
	} else {
		c.l.M("Invalid hostname:", name)
	}
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
