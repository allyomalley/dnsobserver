package main

import (
	"flag"
	"fmt"
	"github.com/miekg/dns"
	"github.com/slack-go/slack"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"strings"
	"time"
)

type Config struct {
	Domain       string `yaml:"domain"`
	PublicIP     string `yaml:"public_ip"`
	SlackWebhook string `yaml:"webhook"`
	Records      []RR   `yaml:"a_records"`
}

type CustomRecords struct {
	Records []RR `yaml:"a_records"`
}

type RR struct {
	Hostname string `yaml:"hostname"`
	IP       string `yaml:"ip"`
}

var conf Config
var answersMap map[string]string

func sendSlack(message string) {
	msg := slack.WebhookMessage{
		Text: message,
	}
	_ = slack.PostWebhook(conf.SlackWebhook, &msg)
}

func handleInteraction(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)
	remoteAddr := w.RemoteAddr().String()
	q1 := r.Question[0]
	t := time.Now()

	if dns.IsSubDomain(q1.Name, conf.Domain+".") && q1.Name != "ns1."+conf.Domain+"." && q1.Name != "ns2."+conf.Domain+"." {
		addrParts := strings.Split(remoteAddr, ":")
		dateString := "Received at: " + "`" + t.Format("January 2, 2006 3:04 PM") + "`"
		fromString := "Received From: " + "`" + addrParts[0] + "`"
		nameString := "Lookup Query: " + "`" + q1.Name + "`"
		typeString := "Query Type: " + "`" + dns.TypeToString[q1.Qtype] + "`"

		message := "*Received DNS interaction:*" + "\n\n" + dateString + "\n" + fromString + "\n" + nameString + "\n" + typeString
		if conf.SlackWebhook != "" {
			sendSlack(message)
		} else {
			fmt.Println(message)
		}
	}

	switch r.Question[0].Qtype {
	case dns.TypeA:
		msg.Authoritative = true
		domain := msg.Question[0].Name
		address, ok := answersMap[domain]
		if ok {
			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 1},
				A:   net.ParseIP(address),
			})
		}
	}
	w.WriteMsg(&msg)
}

func loadConfig() {
	domain := flag.String("domain", "", "Your registered domain name")
	ip := flag.String("ip", "", "Your server's public IP address")
	webhook := flag.String("webhook", "", "Your Slack webhook URL")
	recordsPath := flag.String("recordsFile", "", "Optional path to custom records config file")
	flag.Parse()

	conf = Config{Domain: *domain, PublicIP: *ip, SlackWebhook: *webhook}
	if *recordsPath != "" {
		recs := CustomRecords{}
		data, err := ioutil.ReadFile(*recordsPath)
		if err != nil {
			panic(err)
		}
		err = yaml.Unmarshal(data, &recs)
		if err != nil {
			panic(err)
		}
		conf.Records = recs.Records
	}

	answersMap = map[string]string{
		conf.Domain + ".":          conf.PublicIP,
		"ns1." + conf.Domain + ".": conf.PublicIP,
		"ns2." + conf.Domain + ".": conf.PublicIP,
	}

	for _, record := range conf.Records {
		answersMap[record.Hostname+"."] = record.IP
	}
}

func main() {
	fmt.Println("Configuring...")
	loadConfig()
	if conf.Domain == "" || conf.PublicIP == "" {
		fmt.Println("Error: Must supply a domain and public IP in config file")
		return
	} else {
		fmt.Println("Listener starting!")
	}

	dns.HandleFunc(".", handleInteraction)
	if err := dns.ListenAndServe(conf.PublicIP+":53", "udp", nil); err != nil {
		fmt.Println(err.Error())
		return
	}
}
