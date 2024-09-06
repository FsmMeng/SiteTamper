package core

import (
	"bufio"
	"crypto/tls"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type UrlResult struct {
	Url        string `json:"url"`
	Domain     string `json:"domain"`
	StatusCode int    `json:"status_code"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	IP         string `json:"ip"`
}

func determiEncoding(r *bufio.Reader) encoding.Encoding {
	bytes, err := r.Peek(1024)
	if err != nil {
		return unicode.UTF8
	}
	e, _, _ := charset.DetermineEncoding(bytes, "")
	return e
}

func SimpleReq(url string) UrlResult {
	var urlResult UrlResult
	urlResult.Url = url
	_, domain, subDomain, ips := ParseURL(url)
	parseIP := net.ParseIP(subDomain)
	if parseIP != nil {
		domain = parseIP.String()
	}
	urlResult.Domain = domain

	if len(ips) > 0 {
		urlResult.IP = ips[0]
	}

	transport := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transport, Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return urlResult
	}

	req.Header.Set("User-Agent", UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return urlResult
	}
	defer resp.Body.Close()

	urlResult.StatusCode = resp.StatusCode
	bodyReader := bufio.NewReader(resp.Body)
	e := determiEncoding(bodyReader)
	utf8Reader := transform.NewReader(bodyReader, e.NewDecoder())
	body, err := io.ReadAll(utf8Reader)
	if err != nil {
		urlResult.Content = ""
	} else {
		urlResult.Content = string(body)
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		urlResult.Title = ""
	} else {
		urlResult.Title = doc.Find("title").Text()
	}
	return urlResult
}
