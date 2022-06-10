package cache

import (
	"fmt"
	"sync"

	"github.com/kwitsch/OmadaSiteDns/util"
)

type Cache struct {
	dns  map[string]string
	rdns map[string]string
	lock sync.RWMutex
}

func New() *Cache {
	return &Cache{
		dns:  make(map[string]string),
		rdns: make(map[string]string),
	}
}

func (c *Cache) Update(hostname string, ips ...string) {
	if len(ips) > 0 {
		c.lock.Lock()
		c.dns[hostname] = ips[0]
		c.lock.Unlock()
		for _, i := range ips {
			c.addRDns(hostname, i)
		}
	}
}

func (c *Cache) GetIp(hostname string) (string, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	ip, ok := c.dns[hostname]
	return ip, ok
}

func (c *Cache) GetHostname(reverseIP string) (string, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	hostname, ok := c.rdns[reverseIP]
	return hostname, ok
}

func (c *Cache) Print() {
	fmt.Println("dns cache:")
	for n, v := range c.dns {
		fmt.Println("-", n, "=", v)
	}
	fmt.Println("rdns cache:")
	for n, v := range c.rdns {
		fmt.Println("-", n, "=", v)
	}
}

func (c *Cache) addRDns(hostname, ip string) {
	if revIp, revErr := util.ReverseIP(ip); revErr == nil {
		c.lock.Lock()
		defer c.lock.Unlock()
		c.rdns[revIp] = hostname
	}
}
