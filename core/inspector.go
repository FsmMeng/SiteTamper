package core

import (
	"github.com/antlabs/strsim"
	"regexp"
	"strings"
)

func Inspector(rules []Rule, content string, title string) []map[string]interface{} {
	var defaultEventData []map[string]interface{}
	for _, ruleModel := range rules {
		groupName := strings.TrimSpace(ruleModel.GroupName)
		rule := strings.TrimSpace(ruleModel.Content)
		location := strings.TrimSpace(ruleModel.Location)

		if location == "title" {
			if rule == "" || title == "" {
				return defaultEventData
			} else {
				isMatch, eventData := matchText(rule, groupName, title)
				if isMatch {
					return eventData
				}
			}
		} else {
			if rule == "" || content == "" {
				return defaultEventData
			} else {
				isMatch, eventData := matchText(rule, groupName, content)
				if isMatch {
					return eventData
				}
			}
		}
	}
	return defaultEventData
}

func matchText(rule string, groupName string, text string) (bool, []map[string]interface{}) {
	var eventData []map[string]interface{}
	var allContext []string
	isMatch := true
	ruleParts := strings.Split(rule, "+")
	for _, subrule := range ruleParts {
		if !strings.HasPrefix(subrule, "(?i)") {
			subrule = "(?i)" + subrule
		}
		re := regexp.MustCompile(subrule)
		matches := re.FindAllStringIndex(text, -1)
		if len(matches) > 0 {
			match := matches[0]
			start, end := match[0], match[1]
			contextStart := max(0, start-30)
			contextEnd := min(len(text), end+30)
			context := text[contextStart:contextEnd]

			if len(allContext) > 0 {
				include := false
				for _, i := range allContext {
					include = strings.Contains(i, context)
					if !include {
						include = strings.Contains(context, i)
					}
					if include {
						break
					}
				}

				matchContext := strsim.FindBestMatchOne(context, allContext)
				if matchContext.Score < 0.7 && !include {
					allContext = append(allContext, context)
				}
			} else {
				allContext = append(allContext, context)
			}
		} else {
			isMatch = false
		}
	}
	if isMatch {
		var allContextStr string
		for _, context := range allContext {
			if allContextStr == "" {
				allContextStr = context
			} else {
				allContextStr = allContextStr + " || " + context
			}
		}
		eventData = append(eventData, map[string]interface{}{
			"rule":    rule,
			"group":   groupName,
			"context": allContextStr,
		})
	}
	return isMatch, eventData
}
