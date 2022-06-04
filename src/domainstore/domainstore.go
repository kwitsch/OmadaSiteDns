package domainstore

import (
	"net"
	"strings"
	"sync"

	"github.com/kwitsch/omadaclient/log"
)

type Domainstore struct {
	networks  map[string]network
	overrides map[string]string
	lock      sync.RWMutex
	l         *log.Log
}

type network struct {
	subnet *net.IPNet
	domain string
}

func New(verbose bool) *Domainstore {
	res := Domainstore{
		networks:  make(map[string]network),
		overrides: make(map[string]string),
		l:         log.New("Domainstore", verbose),
	}
	return &res
}

func (ds *Domainstore) AddOverride(name, domain string) {
	ds.lock.Lock()
	defer ds.lock.Unlock()

	iname := strings.ToLower(name)

	ds.l.V("Add override:", iname, domain)

	ds.overrides[iname] = domain
}

func (ds *Domainstore) RemoveOverride(name string) {
	ds.lock.Lock()
	defer ds.lock.Unlock()

	iname := strings.ToLower(name)
	if _, ok := ds.overrides[iname]; ok {
		delete(ds.overrides, iname)
	}
}

func (ds *Domainstore) AddNetwork(name, subnet, domain string) error {
	ds.lock.Lock()
	defer ds.lock.Unlock()

	iname := strings.ToLower(name)
	tdom := domain
	if _, ok := ds.overrides[iname]; ok {
		tdom = ds.overrides[iname]
	}

	_, snet, err := net.ParseCIDR(subnet)
	if err != nil {
		return err
	}

	ds.l.V("Add network:", iname, subnet, tdom)

	ds.networks[iname] = network{
		subnet: snet,
		domain: tdom,
	}

	return nil
}

func (ds *Domainstore) RemoveNetwork(name string) {
	ds.lock.Lock()
	defer ds.lock.Unlock()

	iname := strings.ToLower(name)
	if _, ok := ds.networks[iname]; ok {
		delete(ds.networks, iname)
	}
}

func (ds *Domainstore) GetDomain(ipaddr string) (string, bool) {
	ds.lock.RLock()
	defer ds.lock.RUnlock()

	ip := net.ParseIP(ipaddr)

	for _, v := range ds.networks {
		if v.subnet.Contains(ip) {
			return v.domain, true
		}
	}

	return "", false
}
