package core

import (
	"fmt"
	"github.com/panjf2000/ants/v2"
	"github.com/xuri/excelize/v2"
	"strings"
	"sync"
)

type Rule struct {
	GroupName string
	Content   string
	Location  string
}

func GetCheckRules() []Rule {
	rules := make([]Rule, 0)
	f, err := excelize.OpenFile("core/rules.xlsx")
	if err != nil {
		fmt.Println(err)
		return rules
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		fmt.Println(err)
		return rules
	}
	for i, row := range rows {
		if i == 0 {
			continue
		}

		rule := Rule{
			GroupName: row[0],
			Content:   row[1],
			Location:  row[2],
		}
		rules = append(rules, rule)
	}
	return rules
}

func Checker(links []string) {
	rules := GetCheckRules()

	var wg sync.WaitGroup
	pool, _ := ants.NewPoolWithFunc(10, func(i interface{}) {
		argMap := i.(map[string]interface{})
		check(argMap)
		wg.Done()
	})
	defer pool.Release()

	for _, link := range links {
		wg.Add(1)
		argMap := map[string]interface{}{
			"link":  link,
			"rules": rules,
		}
		_ = pool.Invoke(argMap)
	}
	wg.Wait()
}

func check(argMap map[string]interface{}) {
	link := argMap["link"].(string)
	rules := argMap["rules"].([]Rule)
	var info string

	// get page
	urlResult := SimpleReq(link)
	if urlResult.StatusCode == 0 {
		urlResult = SimpleReq(link)
	}
	if urlResult.StatusCode == 0 {
		fmt.Printf("【Failed】 %s\n", link)
		return
	}
	info = fmt.Sprintf("【url】    : %s\n【domain】 : %s\n【title】  : %s\n", urlResult.Url, urlResult.Domain, urlResult.Title)

	// inspect
	eventData := Inspector(rules, urlResult.Content, urlResult.Title)
	if len(eventData) == 0 {
		fmt.Printf("【healthy】 %s\n", link)
		return
	}
	eventDetail := CleanInvalidUTF8(eventData[0]["context"].(string))
	eventDetail = strings.ReplaceAll(eventDetail, "\r", " ")
	eventDetail = strings.ReplaceAll(eventDetail, "\n", " ")
	info = fmt.Sprintf("%s【group】  : %s\n【context】: %s\n", info, eventData[0]["group"], eventDetail)

	// get icp
	icpInfo := GetIcp(urlResult.Domain)
	if icpInfo.IcpNumber != "" {
		info = fmt.Sprintf("%s【unit】   : %s\n【permit】 : %s\n【nature】 : %s\n", info, icpInfo.UnitName, icpInfo.IcpNumber, icpInfo.Nature)
	}

	equals := strings.Repeat("=", 100)
	fmt.Printf("\n检测结果如下：\n%s\n%v%s\n", equals, info, equals)
}
