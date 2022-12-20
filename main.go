package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	crtURL = "https://crt.sh"
)

type record struct {
	IssuerCAID int    `json:"issuer_ca_id"`
	IssuerName string `json:"issuer_name"`
	CommonName string `json:"common_name"`
	NameValue  string `json:"name_value"`
	ID         int    `json:"id"`
}

type result struct {
	hostname string
	ips      []string
}

func main() {
	checkLive := flag.Bool("l", false, "check whether host is live")
	dns := flag.String("d", "8.8.8.8", "dns server to use")
	port := flag.Int("p", 53, "dns port number")
	flag.Parse()

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: crtsh [-l=true|false] domain.com")
		os.Exit(1)
	}

	if len(flag.Args()) != 1 {
		flag.Usage()
	}

	_ = checkLive

	domain := flag.Args()[0]

	req, err := http.NewRequest("GET", crtURL, nil)
	if err != nil {
		panic(err)
	}

	q := req.URL.Query()
	q.Add("q", domain)
	q.Add("output", "json")
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var records []record
	if err := json.NewDecoder(resp.Body).Decode(&records); err != nil {
		panic(err)
	}

	names := make(map[string]struct{})

	for _, rec := range records {
		values := strings.Split(rec.NameValue, "\n")
		for _, val := range values {
			if val[0] != '*' {
				names[val] = struct{}{}
			}
		}
	}

	if *checkLive {
		r := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Second * time.Duration(5),
				}
				return d.DialContext(ctx, network, fmt.Sprintf("%s:%d", *dns, *port))
			},
		}

		var results []*result

		ch := make(chan string, 10)
		res := make(chan *result, 10)
		var wg sync.WaitGroup
		wg.Add(10)

		go func() {
			for r := range res {
				results = append(results, r)
			}
		}()

		for i := 0; i < 10; i++ {
			go check(r, ch, res, &wg)
		}

		for name := range names {
			ch <- name
		}
		close(ch)

		go func() {
			wg.Wait()
			close(res)
		}()

		countUnresolvable := 0
		for _, r := range results {
			if len(r.ips) == 1 && r.ips[0] == "unresolvable" {
				countUnresolvable++
			}
		}

		fmt.Printf("[*] Fetched domains for %q (%d resolvable; %d unresolvable\n",
			domain, len(results)-countUnresolvable, countUnresolvable)
		for _, r := range results {
			fmt.Printf("[*] %s => %s\n", r.hostname, strings.Join(r.ips, ", "))
		}
	}
}

func check(r *net.Resolver, ch <-chan string, res chan<- *result, wg *sync.WaitGroup) {
	defer wg.Done()
	for c := range ch {
		resolved := &result{}
		resolved.hostname = c
		ips, err := r.LookupIP(context.Background(), "ip4", c)
		if err != nil {
			resolved.ips = []string{"unresolvable"}
		} else {
			resolved.ips = getIps(ips)
		}
		res <- resolved
	}
}

func getIps(ips []net.IP) []string {
	res := make([]string, len(ips))
	for i, ip := range ips {
		res[i] = ip.String()
	}
	return res
}
