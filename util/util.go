package util

import (
	"fmt"
	"strings"

	"github.com/kwitsch/omadaclient/utils"
	"github.com/miekg/dns"
	"golang.org/x/net/publicsuffix"
)

func ReverseIP(ip string) (string, error) {
	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		return fmt.Sprintf("%s.%s.%s.%s", parts[3], parts[2], parts[1], parts[0]), nil
	}
	return "", utils.NewError("invalid ip")
}

func QName(req *dns.Msg) string {
	return strings.TrimSuffix(strings.ToLower(req.Question[0].Name), ".")
}

func AnswerToString(answer []dns.RR) string {
	answers := make([]string, len(answer))

	for i, record := range answer {
		switch v := record.(type) {
		case *dns.A:
			answers[i] = fmt.Sprintf("A (%s)", v.A)
		case *dns.AAAA:
			answers[i] = fmt.Sprintf("AAAA (%s)", v.AAAA)
		case *dns.CNAME:
			answers[i] = fmt.Sprintf("CNAME (%s)", v.Target)
		case *dns.PTR:
			answers[i] = fmt.Sprintf("PTR (%s)", v.Ptr)
		default:
			answers[i] = fmt.Sprint(record.String())
		}
	}

	return strings.Join(answers, ", ")
}

func TLDPlusOne(req *dns.Msg) string {
	res, _ := publicsuffix.EffectiveTLDPlusOne(QName(req))
	return res
}
