package server

import (
	"fmt"
	"net"
	"strings"

	"github.com/kwitsch/OmadaSiteDns/cache"
	"github.com/kwitsch/OmadaSiteDns/config"
	"github.com/kwitsch/omadaclient/log"
	"github.com/miekg/dns"
)

type Server struct {
	udp   *dns.Server
	tcp   *dns.Server
	cache *cache.Cache
	ttl   uint32
	l     *log.Log
	Error chan (error)
}

func New(cache *cache.Cache, cfg config.Server, verbose bool) *Server {
	res := &Server{
		udp:   createUDPServer(),
		tcp:   createTCPServer(),
		cache: cache,
		ttl:   uint32(cfg.Ttl),
		l:     log.New("Server", verbose),
		Error: make(chan error, 2),
	}

	res.setupHandlers()

	return res
}

func (s *Server) Start() {
	go func() {
		s.Error <- s.udp.ListenAndServe()
	}()
	go func() {
		s.Error <- s.tcp.ListenAndServe()
	}()
}

func (s *Server) Stop() {
	s.udp.Shutdown()
	s.tcp.Shutdown()
}

func (s *Server) setupHandlers() {
	uh := s.udp.Handler.(*dns.ServeMux)
	uh.HandleFunc(".", s.OnRequest)

	th := s.tcp.Handler.(*dns.ServeMux)
	th.HandleFunc(".", s.OnRequest)
}

func createUDPServer() *dns.Server {
	return &dns.Server{
		Addr:    ":53",
		Net:     "udp",
		Handler: dns.NewServeMux(),
		NotifyStartedFunc: func() {
			fmt.Println("UDP server is up and running")
		},
		UDPSize: 65535,
	}
}

func createTCPServer() *dns.Server {
	return &dns.Server{
		Addr:    ":53",
		Net:     "tcp",
		Handler: dns.NewServeMux(),
		NotifyStartedFunc: func() {
			fmt.Println("TCP server is up and running")
		},
	}
}

const rdnsSuf string = ".in-addr.arpa"

func (s *Server) OnRequest(w dns.ResponseWriter, request *dns.Msg) {
	q := request.Question[0]
	s.l.V("Requst:", q.Name, "Type:", q.Qtype)
	m := new(dns.Msg)
	m.SetReply(request)

	if q.Qtype == dns.TypePTR || q.Qtype == dns.TypeA {
		cname := strings.TrimSuffix(strings.ToLower(q.Name), ".")
		exists := false
		val := ""

		if q.Qtype == dns.TypePTR {
			crname := strings.TrimSuffix(cname, rdnsSuf)

			val, exists = s.cache.GetHostname(crname)
			if exists {
				rr := new(dns.PTR)
				rr.Hdr = dns.RR_Header{
					Name:   q.Name,
					Rrtype: dns.TypePTR,
					Class:  dns.ClassINET,
					Ttl:    s.ttl,
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
					Ttl:    s.ttl,
				}

				rr.A = net.ParseIP(val)

				m.Answer = []dns.RR{rr}

				s.l.Return(val)
			}
		}

		if !exists {
			s.l.Return("NXDomain")
			m.SetRcode(request, dns.RcodeNameError)
		}
	}
	w.WriteMsg(m)
}
