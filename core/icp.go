package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/dop251/goja"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type ChinazIcpModel struct {
	Domain     string     `json:"domain"`
	UnitName   string     `json:"unit_name"`
	IcpNumber  string     `json:"icp_number"`
	Nature     string     `json:"nature"`
	Title      string     `json:"title"`
	ReviewDate *time.Time `json:"review_date"`
}

func getChinazToken(argsStr string, timestamp int64) (string, error) {
	jsCode, err := os.ReadFile("core/chinaz_token.js")
	if err != nil {
		return "", err
	}

	vm := goja.New()
	_, err = vm.RunString(string(jsCode))
	if err != nil {
		fmt.Println("JS error")
		return "", err
	}
	getToken, ok := goja.AssertFunction(vm.Get("get_token"))
	if !ok {
		return "", err
	}

	result, err := getToken(goja.Undefined(), vm.ToValue(argsStr), vm.ToValue(timestamp))
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

func GetIcp(domain string) ChinazIcpModel {
	var icpResp *http.Response
	var err error

	/*
		ICP Base
	*/

	for i := 0; i < 3; i++ {
		transport := &http.Transport{}
		httpClient := &http.Client{Transport: transport, Timeout: 15 * time.Second}

		req, err := http.NewRequest("GET", fmt.Sprintf("https://icp.chinaz.com/%s", domain), nil)
		req.Header.Set("user-agent'", UserAgent)
		icpResp, err = httpClient.Do(req)

		if err != nil {
			continue
		}

		break
	}
	if icpResp == nil {
		return ChinazIcpModel{}
	}
	defer icpResp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(icpResp.Body)
	if err != nil {
		fmt.Printf("parse icp error %s | %s\n", domain, err)
		return ChinazIcpModel{}
	}
	icpInfo := ChinazIcpModel{Domain: domain}
	icpInfo.UnitName = doc.Find("#companyName").Text()
	icpInfo.Nature = doc.Find("#icp_content > div:nth-child(1) > div:nth-child(2) > div > table > tbody > tr:nth-child(1) > td:nth-child(4)").Text()
	reviewDate := doc.Find("#icp_content > div:nth-child(1) > div:nth-child(2) > div > table > tbody > tr:nth-child(2) > td:nth-child(4)").Text()
	icpInfo.Title = doc.Find("#icp_content > div:nth-child(1) > div:nth-child(2) > div > table > tbody > tr:nth-child(3) > td:nth-child(2)").Text()

	if icpInfo.UnitName == "" || icpInfo.Nature == "个人" {
		return ChinazIcpModel{}
	}

	reviewDateT, _ := time.Parse(TimeFormatSecond, reviewDate)
	icpInfo.ReviewDate = &reviewDateT

	/*
		ICP Number
	*/

	payload := make(map[string]interface{})
	payload["keyword"] = domain
	jsonBuf, _ := json.Marshal(payload)
	timestamp := time.Now().UnixNano() / 1000000
	tokenStr, err := getChinazToken(string(jsonBuf), timestamp)
	if tokenStr == "" || err != nil {
		fmt.Printf("get chinaz token error %s | %s\n", domain, err)
		return ChinazIcpModel{}
	}
	tokenParts := strings.Split(tokenStr, "#")
	if len(tokenParts) != 2 {
		fmt.Printf("parse chinaz token error %s | %s\n", domain, err)
		return ChinazIcpModel{}
	}
	token := tokenParts[0]
	key := tokenParts[1]
	rd := strings.Split(key, ",")[0]

	permitReq, _ := http.NewRequest("POST", "https://icp.chinaz.com/index/api/queryPermit", bytes.NewBuffer(jsonBuf))
	permitReq.Header.Set("Content-Type", "application/json")
	permitReq.Header.Set("rd", rd)
	permitReq.Header.Set("token", token)
	permitReq.Header.Set("key", key)
	permitReq.Header.Set("ts", strconv.FormatInt(timestamp, 10))
	permitReq.Header.Set("user-agent'", UserAgent)

	var icpNumber string
	var permitResp *http.Response
	for i := 0; i < 3; i++ {
		transport := &http.Transport{}
		httpClient := &http.Client{Transport: transport, Timeout: 15 * time.Second}
		permitResp, err = httpClient.Do(permitReq)
		if err != nil {
			continue
		}

		var result map[string]interface{}
		json.NewDecoder(permitResp.Body).Decode(&result)

		if data, ok := result["data"].(map[string]interface{}); ok && data != nil {
			if permit, ok := data["permit"].(string); ok {
				icpNumber = permit
			} else {
				continue
			}
		} else {
			continue
		}

		break
	}
	if permitResp == nil {
		return ChinazIcpModel{}
	}

	defer permitResp.Body.Close()
	if icpNumber == "" {
		return ChinazIcpModel{}
	}

	icpInfo.IcpNumber = icpNumber
	return icpInfo
}
