package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

var jar, _ = cookiejar.New(nil)

type Header http.Header

func (h *Header) String() string {
	return fmt.Sprint(*h)
}

func (h *Header) Set(value string) error {
	slice := strings.SplitN(value, ":", 2)
	if len(slice) != 2 {
		return errors.New("Invalid header format")
	}
	key := strings.TrimSpace(slice[0])
	http.Header(*h).Add(key, strings.TrimSpace(slice[1]))
	return nil
}

func main() {
	var compressed bool
	var quiet bool
	var data string
	selector := "body"
	method := "GET"
	interval := int64(30)
	header := make(Header)
	flagSet := flag.NewFlagSet("", flag.ExitOnError)
	flagSet.Var(&header, "H", "HTTP header")
	flagSet.StringVar(&method, "X", method, "GET or POST")
	flagSet.BoolVar(&compressed, "compressed", true, "enable compression")
	flagSet.StringVar(&data, "data", "", "POST data")
	flagSet.Int64Var(&interval, "interval", interval, "Pull data every N seconds")
	flagSet.BoolVar(&quiet, "q", false, "quiet mode")
	flagSet.StringVar(&selector, "selector", selector, "jQuerish element selector")
	if len(os.Args) < 2 {
		flagSet.PrintDefaults()
		return
	}
	urlStr := os.Args[1]
	if strings.HasPrefix(urlStr, "-") {
		flagSet.Parse(os.Args[1:])
		urlStr = flagSet.Arg(0)
	} else {
		flagSet.Parse(os.Args[2:])
	}
	if data != "" && method == "GET" {
		method = "POST"
	}
	doc, err := request(method, urlStr, data, header)
	if err != nil {
		log.Fatalln(err)
	}
	text := extract(doc.Find(selector))
	if !quiet {
		log.Println(text)
	}
	for {
		time.Sleep(time.Second * time.Duration(interval))
		doc, err := request(method, urlStr, data, header)
		if err != nil {
			log.Fatalln(err)
		}
		newText := extract(doc.Find(selector))
		if newText != text {
			if !quiet {
				log.Println("something changed")
				fmt.Println(newText)
			}
			return
		}
		log.Println("nothing changed")
	}
}

func request(method, urlStr, data string, header Header) (*goquery.Document, error) {
	client := &http.Client{Jar: jar}
	body := strings.NewReader(data)
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	req.Header = http.Header(header)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return goquery.NewDocumentFromResponse(resp)
}

func extract(s *goquery.Selection) string {
	return strings.Map(noSpace, s.Text())
}

func noSpace(r rune) rune {
	if unicode.IsSpace(r) {
		return -1
	}
	return r
}
