package querylogger

import (
	"os"
	"time"

	"github.com/kwitsch/OmadaSiteDns/cache"
	"github.com/kwitsch/OmadaSiteDns/config"
	"github.com/kwitsch/OmadaSiteDns/util"
	"github.com/kwitsch/omadaclient/log"
	"github.com/miekg/dns"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type LogEntry struct {
	ClientIp string
	Request  *dns.Msg
	Response *dns.Msg
	Start    time.Time
	Duration int64
}

type QueryLogger struct {
	cfg    *config.Logger
	l      *log.Log
	client influxdb2.Client
	cache  *cache.Cache
	ichan  chan LogEntry
	schan  chan bool
}

func New(cfg config.Logger, cache *cache.Cache, verbose bool) *QueryLogger {
	res := QueryLogger{
		cfg:   &cfg,
		l:     log.New("QueryLogger", verbose),
		cache: cache,
	}

	if res.cfg.Enabled {
		res.client = influxdb2.NewClient(cfg.Url, cfg.Token)
		res.ichan = make(chan LogEntry)
		res.schan = make(chan bool)
	}

	return &res
}

func (ql *QueryLogger) Start() {
	if ql.cfg.Enabled {
		hname := os.Getenv("HOSTNAME")
		writeAPI := ql.client.WriteAPI(ql.cfg.Org, ql.cfg.Bucket)

		go func() {
			for {
				select {
				case m := <-ql.ichan:
					p := influxdb2.NewPoint(hname, map[string]string{}, m.String(ql.cache), m.Start)
					writeAPI.WritePoint(p)
				case e := <-writeAPI.Errors():
					ql.l.E(e)
				case <-ql.schan:
					return
				}
			}
		}()
	} else {
		ql.l.M("is disabled")
	}
}

func (ql *QueryLogger) Close() {
	if ql.cfg.Enabled {
		ql.schan <- true
		close(ql.ichan)
		close(ql.schan)
		ql.client.Close()
	}
}

func (ql *QueryLogger) Log(le LogEntry) {
	if ql.cfg.Enabled {
		ql.ichan <- le
	}
}

func (le *LogEntry) String(cache *cache.Cache) map[string]interface{} {
	res := map[string]interface{}{
		"ClientIP":      le.ClientIp,
		"ClientName":    le.QName(cache),
		"DurationMs":    le.Duration,
		"QuestionType":  dns.TypeToString[le.Request.Question[0].Qtype],
		"QuestionName":  util.QName(le.Request),
		"EffectiveTLDP": util.TLDPlusOne(le.Request),
		"Answer":        util.AnswerToString(le.Response.Answer),
		"ResponseCode":  dns.RcodeToString[le.Response.Rcode],
	}

	return res
}

func (le *LogEntry) QName(cache *cache.Cache) string {
	revip, reverr := util.ReverseIP(le.ClientIp)
	res := le.ClientIp
	if reverr == nil {
		res, _ = cache.GetHostname(revip)
	}
	return res
}