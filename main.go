package main

import (
	"bytes"
	"crypto/sha1"
	"crypto/tls"
	"fmt"
	"github.com/dgraph-io/badger/badger"
	"github.com/jackdanger/collectlinks"
	"github.com/temoto/robotstxt-go"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"
)

var user_agent string = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.152 Safari/537.36"
var client *http.Client
var last_fetch string = ""
var crawl_delay = 250 * time.Millisecond

var current_id uint64 = 0
var first_id uint64 = 0

var db_path string = "/tmp/nuspider_db"
var kv *badger.KV

var site string

var site_uri *url.URL
var robots *robotstxt.RobotsData

type PageJson struct {
	Url  string `json:"url"`
	Html string `json:"html"`
}

func push(val string) {
	key := []byte(fmt.Sprintf("q_%d", current_id))
	kv.Set(key, []byte(val))

	current_id += 1
}

func pop() string {
	pop_key := []byte(fmt.Sprintf("q_%d", first_id))
	val, _ := kv.Get(pop_key)

	first_id += 1

	return string(val)
}

func hashUrl(u string) string {
	h := sha1.New()
	io.WriteString(h, u)

	return fmt.Sprintf("%x", h.Sum(nil))
}

func hasVisited(u string) bool {
	val, _ := kv.Get([]byte(hashUrl(u)))

	if val != nil {
		return true
	} else {
		return false
	}
}

func setVisited(u string) {
	kv.Set([]byte(hashUrl(u)), []byte("1"))
}

func fixUrl(href string) *url.URL {
	// for links such as <a href="//blah.domain.com/abc.html">
	if strings.HasPrefix(href, "//") {
		return nil
	}

	// for links such as <a href="/abc.html">
	if strings.HasPrefix(href, "/") {
		href = fmt.Sprint(site_uri.Scheme, "://", site_uri.Host, href)
	}

	// for links such as <a href="abc.html">
	if !strings.HasPrefix(href, "/") && !strings.HasPrefix(href, "http") {
		href = fmt.Sprint(site_uri.Scheme, "://", site_uri.Host, "/", href)
	}

	uri, err := url.Parse(href)
	if err != nil {
		return nil
	}

	//ignore interpage links
	if uri.Fragment != "" {
		return nil
	}

	// return if its not the same host
	if uri.Host != site_uri.Host {
		return nil
	}

	return uri
}

func fetch(u string) {
	log.Println("Fetching", u)

	uri, _ := url.Parse(u)

	req, err := http.NewRequest("GET", u, nil)
	req.Header.Add("User-Agent", user_agent)
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Add("Accept-Language", "en-US,en;q=0.8,fr;q=0.6")
	req.Header.Add("Cache-Control", "max-age=0")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Host", uri.Host)

	if last_fetch != "" {
		req.Header.Add("Referer", u)
	}

	resp, err := client.Do(req)

	last_fetch = u

	if err != nil {
		fmt.Printf("FETCH ERROR: %v - %s\n", err, u)
		return
	} else if resp.StatusCode != 200 {
		if resp.Body != nil {
			resp.Body.Close()
		}

		fmt.Printf("FETCH ERROR: Status %d - %s\n", resp.StatusCode, u)
		return
	}

	if site_uri == nil {
		site_uri = resp.Request.URL
	}

	setVisited(u)

	// need to copy the resp.Body because you can
	// only read the bytes once
	// keep this as a place holder for when we want to do something with it
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	// bodyStr := string(bodyBytes)
	links := collectlinks.All(resp.Body)

	resp.Body.Close()

	for _, link := range links {
		page_url := fixUrl(link)

		if page_url != nil && (robots == nil || (robots != nil && robots.TestAgent(page_url.Path, "*"))) {
			urlStr := page_url.String()

			if !hasVisited(urlStr) && urlStr != "" {
				push(urlStr)
				setVisited(urlStr)
			}
		}
	}

	/* Do something with it here */
}

func main() {
	log.SetOutput(os.Stdout)

	if len(os.Args) != 2 {
		fmt.Println("Usage: ./nuspider <site>")
		os.Exit(0)
	} else {
		site = os.Args[1]
	}

	log.Println("Starting crawler for", site)

	os.RemoveAll(db_path)
	os.MkdirAll(db_path, 0700)

	opt := badger.DefaultOptions
	opt.Dir = db_path
	kv, _ = badger.NewKV(&opt)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	cookieJar, _ := cookiejar.New(nil)

	client = &http.Client{Transport: transport, Jar: cookieJar}

	resp, err := http.Get(fmt.Sprint("http://", site, "/robots.txt"))

	if err != nil {
		robots = nil
	} else {
		robots, _ = robotstxt.FromResponse(resp)
		resp.Body.Close()
	}

	push(fmt.Sprint("http://", site, "/"))

	for {
		fetch(pop())
		time.Sleep(crawl_delay)

		if first_id%100 == 0 {
			log.Println("Pages crawled: ", first_id+1)
			log.Println("Queue length: ", current_id-first_id)
		}
	}
}
