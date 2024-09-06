package core

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/bobesa/go-domain-util/domainutil"
	"net"
	"net/url"
	"strings"
	"unicode/utf8"
)

const (
	UserAgent        = "Mozilla/5.0 (compatible; Baiduspider/2.0;+http://www.baidu.com/search/spider.html)"
	TimeFormatSecond = "2006-01-02 15:04:05"
)

func ParseURL(inputURL string) (string, string, string, []string) {
	validURL := govalidator.IsURL(inputURL)
	if !validURL {
		return "", "", "", nil
	}
	parsedURL, err := url.Parse(inputURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return "", "", "", nil
	}
	parts := strings.Split(parsedURL.Hostname(), ".")
	if len(parts) < 2 {
		return "", "", "", nil
	}

	subDomain := parsedURL.Hostname()
	domain := domainutil.Domain(parsedURL.Hostname())

	var website string
	if parsedURL.Port() == "80" || parsedURL.Port() == "443" || parsedURL.Port() == "" {
		website = fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Hostname())
	} else {
		website = fmt.Sprintf("%s://%s:%s", parsedURL.Scheme, parsedURL.Hostname(), parsedURL.Port())
	}
	ipList, err := net.LookupIP(parsedURL.Hostname())
	if err != nil {
		return website, domain, subDomain, nil
	}
	ips := make([]string, len(ipList))
	for i, ip := range ipList {
		ips[i] = ip.String()
	}

	return website, domain, subDomain, ips
}

func CleanInvalidUTF8(s string) string {
	var result []rune
	for _, r := range s {
		if r == utf8.RuneError {
			continue // 跳过非法的 UTF-8 字节序列
		}
		result = append(result, r)
	}
	return string(result)
}
