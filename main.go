package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
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

func main() {
	req, err := http.NewRequest("GET", crtURL, nil)
	if err != nil {
		panic(err)
	}

	q := req.URL.Query()
	q.Add("q", os.Args[1])
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

	for name := range names {
		fmt.Println(name)
	}

	/*ch := make(chan string, 10)
	var wg sync.WaitGroup
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go checkLive(ch, &wg)
	}

	for name := range names {
		ch <- name
	}
	close(ch)

	wg.Wait()*/
}

func checkLive(ch chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for in := range ch {
		out, err := exec.Command("ping", "-c1", "-w1", in).Output()
		if err != nil && err.Error() != "exit status 68" {
			panic(err)
		}
		if err != nil && err.Error() == "exit status 68" {
			fmt.Printf("%s => unknown\n", in)
		} else {
			ipRex := regexp.MustCompile(`\(.*?\)`)
			matches := ipRex.FindStringSubmatch(string(out))
			if len(matches) > 0 {
				fmt.Printf("%s => %v\n", in, matches[0])
			}
		}
	}
}
