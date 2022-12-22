package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"net"
	"net/http"
	"os"
	"sort"
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
	all := flag.Bool("a", false, "print unresolvable")
	tbl := flag.Bool("t", true, "print results as table")
	dns := flag.String("d", "8.8.8.8", "dns server to use")
	port := flag.Int("p", 53, "dns port number")
	flag.Parse()

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: crtsh [options] domain.com")
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

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
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
			if val[0] != '*' && !strings.Contains(val, "@") {
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

		ch := make(chan string, len(names))
		res := make(chan *result, len(names))
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

		wg.Wait()
		close(res)

		if *tbl {
			printLiveResults(domain, results, *all)
		} else {
			countUnresolvable := 0
			for _, r := range results {
				if len(r.ips) == 1 && r.ips[0] == "unresolvable" {
					countUnresolvable++
				}
			}

			fmt.Printf("[*] Checked live domains for %q - %d (%d resolvable; %d unresolvable)\n",
				domain, len(results), len(results)-countUnresolvable, countUnresolvable)
			for _, r := range results {
				ips := strings.Join(r.ips, ", ")
				if ips == "unresolvable" {
					if *all {
						fmt.Printf("[*] %s => %s\n", r.hostname, ips)
					}
				} else {
					fmt.Printf("[*] %s => %s\n", r.hostname, ips)
				}
			}
		}
	} else {
		if *tbl {
			printBasicTable(domain, names)
		} else {
			fmt.Printf("[*] Extracted data for %q\n", domain)
			for name := range names {
				fmt.Printf("[*] %s\n", name)
			}
		}
	}
}

func printBasicTable(domain string, data map[string]struct{}) {
	i := 0
	domains := make([]string, len(data))
	for name := range data {
		domains[i] = name
		i++
	}

	sort.Strings(domains)
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Name"})

	for i, name := range domains {
		t.AppendRow(table.Row{i + 1, name})
	}

	title := fmt.Sprintf("Extracted domains for %q", domain)
	t.SetTitle(title)
	t.Style().Title.Align = text.AlignCenter

	t.Render()
}

func printLiveResults(domain string, results []*result, all bool) {
	countUnresolvable := 0
	for _, r := range results {
		if len(r.ips) == 1 && r.ips[0] == "unresolvable" {
			countUnresolvable++
		}
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetIndexColumn(1)

	t.AppendHeader(table.Row{"#", "Name", "IP Addresses"})

	for i, r := range results {
		ips := strings.Join(r.ips, ", ")
		if ips == "unresolvable" && !all {
			continue
		} else {
			t.AppendRow(table.Row{i + 1, r.hostname, ips})
		}
	}

	title := fmt.Sprintf("Live domains for %q (%d resolvable; %d unresolvable)", domain,
		len(results)-countUnresolvable, countUnresolvable)
	t.SetTitle(title)
	t.Style().Title.Align = text.AlignCenter

	t.Render()
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
