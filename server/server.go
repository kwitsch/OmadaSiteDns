package server

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/kwitsch/OmadaSiteDns/cache"
	"github.com/kwitsch/OmadaSiteDns/config"
	"github.com/kwitsch/OmadaSiteDns/querylogger"
	"github.com/kwitsch/omadaclient/log"
	"github.com/miekg/dns"
)

const (
	rdnsSuffix    = ".in-addr.arpa"
	serverAddress = ":53"
	udpNet        = "udp"
	tcpNet        = "tcp"
	zeroIP        = "0.0.0.0"
)

type Server struct {
	udp   *dns.Server
	tcp   *dns.Server
	cache *cache.Cache
	cfg   *config.Server
	l     *log.Log
	ql    *querylogger.QueryLogger
	Error chan error
}

func New(cacheIn *cache.Cache, cfgs config.Server, cfgl config.Logger, verbose bool) *Server {
	res := &Server{
		udp:   createUDPServer(),
		tcp:   createTCPServer(),
		cache: cacheIn,
		cfg:   &cfgs,
		l:     log.New("Server", verbose),
		ql:    querylogger.New(cfgl, cacheIn, verbose),
		Error: make(chan error, 2),
	}

	res.setupHandlers()

	return res
}

func (s *Server) Start() {
	s.ql.Start()

	if s.cfg.Udp {
		go func() {
			s.Error <- s.udp.ListenAndServe()
		}()
	}
	if s.cfg.Tcp {
		go func() {
			s.Error <- s.tcp.ListenAndServe()
		}()
	}
}

func (s *Server) Stop() {
	if s.cfg.Udp {
		defer s.udp.Shutdown()
	}
	if s.cfg.Tcp {
		defer s.tcp.Shutdown()
	}
}

func (s *Server) setupHandlers() {
	if s.cfg.Udp {
		uh := s.udp.Handler.(*dns.ServeMux)
		uh.HandleFunc(".", s.OnRequest)
	}

	if s.cfg.Tcp {
		th := s.tcp.Handler.(*dns.ServeMux)
		th.HandleFunc(".", s.OnRequest)
	}
}

func createUDPServer() *dns.Server {
	return &dns.Server{
		Addr:    serverAddress,
		Net:     udpNet,
		Handler: dns.NewServeMux(),
		NotifyStartedFunc: func() {
			fmt.Println("UDP server is up and running")
		},
		UDPSize: 65535,
	}
}

func createTCPServer() *dns.Server {
	return &dns.Server{
		Addr:    serverAddress,
		Net:     tcpNet,
		Handler: dns.NewServeMux(),
		NotifyStartedFunc: func() {
			fmt.Println("TCP server is up and running")
		},
	}
}

func (s *Server) OnRequest(w dns.ResponseWriter, request *dns.Msg) {
	start := time.Now()

	clientip := zeroIP
	if w != nil {
		clientip = resolveClientIP(w.RemoteAddr())
	}
	q := request.Question[0]
	s.l.V("Requst:", q.Name, "Type:", q.Qtype)
	m := new(dns.Msg)
	m.SetReply(request)

	exists := false
	if q.Qtype == dns.TypePTR || q.Qtype == dns.TypeA {
		cname := strings.TrimSuffix(strings.ToLower(q.Name), ".")
		val := ""

		if q.Qtype == dns.TypePTR {
			crname := strings.TrimSuffix(cname, rdnsSuffix)

			val, exists = s.cache.GetHostname(crname)
			if exists {
				rr := new(dns.PTR)
				rr.Hdr = dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypePTR,
					Class:  dns.ClassINET,
					Ttl:    uint32(s.cfg.Ttl),
				}

				rr.Ptr = fmt.Sprintf("%s.", val)

				m.Answer = []dns.RR{rr}

				s.l.Return(val)
			}
		} else if q.Qtype == dns.TypeA {
			val, exists = s.cache.GetIp(cname)
			if exists {
				rr := new(dns.A)
				rr.Hdr = dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    uint32(s.cfg.Ttl),
				}

				rr.A = net.ParseIP(val)

				m.Answer = []dns.RR{rr}

				s.l.Return(val)
			}
		}
	}

	if !exists {
		s.l.Return("NXDomain")
		m.SetRcode(request, dns.RcodeNameError)
	}

	duration := time.Since(start).Milliseconds()

	w.WriteMsg(m)

	s.ql.Log(querylogger.LogEntry{
		ClientIp: clientip,
		Request:  request,
		Response: m,
		Start:    start,
		Duration: duration,
	})
}

func resolveClientIP(addr net.Addr) string {
	if u, ok := addr.(*net.UDPAddr); ok {
		return u.IP.String()
	} else if t, ok := addr.(*net.TCPAddr); ok {
		return t.IP.String()
	}

	return zeroIP
}
