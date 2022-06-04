package config

import (
	"fmt"
	"time"

	"github.com/kwitsch/OmadaSiteDns/osdutils"
	. "github.com/kwitsch/go-dockerutils/config"
	"github.com/kwitsch/omadaclient/utils"
)

type Config struct {
	Verbose int     `koanf:"verbose" default:"0"`
	Site    Site    `koanf:"site"`
	Crawler Crawler `koanf:"crawler"`
	Server  Server  `koanf:"server"`
}

type Site struct {
	Url        string `koanf:"url"`
	Site       string `koanf:"site"`
	Username   string `koanf:"username"`
	Password   string `koanf:"password"`
	SkipVerify bool   `koanf:"skipverify" default:"false"`
}

type Crawler struct {
	HostIntervall    time.Duration `koan:"hostintervall" default:"5m"`
	NetworkIntervall time.Duration `koan:"networkintervall" default:"60m"`
	Converters       map[int]struct {
		Regex      string `koan:"regex"`
		Substitute string `koan:"substitute"`
	} `koan:"converters"`
	Network map[int]struct {
		Name   string `koan:"name"`
		Domain string `koan:"domain"`
	} `koan:"network"`
}

type Server struct {
	Ttl time.Duration `koan:"ttl" default:"5m"`
	Udp bool          `koanf:"udp" default:"true"`
	Tcp bool          `koanf:"tcp" default:"true"`
}

const prefix = "OSD_"

func Get() (*Config, error) {
	var res Config
	if err := Load(prefix, &res); err != nil {
		return nil, err
	}

	if strIsNotSet(res.Site.Url) {
		return nil, utils.NewError("No Omada controller url set")
	}

	if strIsNotSet(res.Site.Username) {
		return nil, utils.NewError("No username set")
	}

	if strIsNotSet(res.Site.Password) {
		return nil, utils.NewError("No password set")
	}

	if !res.Server.Udp && !res.Server.Tcp {
		return nil, utils.NewError("No server enabled")
	}

	for k, v := range res.Crawler.Converters {
		if !osdutils.ValidSubstitute(v.Substitute) {
			return nil, utils.NewError("Converter", "-", k, "has invalid substitute:", v.Substitute)
		}
	}

	for k, v := range res.Crawler.Network {
		if !osdutils.ValidDnsStr(v.Domain) {
			return nil, utils.NewError("Network", "-", k, "("+v.Name+")", "has invalid domain override:", v.Domain)
		}
	}

	if res.Verbose > 1 {
		fmt.Println("Config:", utils.ToString(res))
	}
	return &res, nil
}

func strIsNotSet(input string) bool {
	return (len(input) == 0)
}
