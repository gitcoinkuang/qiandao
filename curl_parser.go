package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	urlRegexp    = regexp.MustCompile(`curl\s+(?:'([^']+)'|"([^"]+)"|([^\s]+))`)
	methodRegexp = regexp.MustCompile(`(?:-X|--request)\s+([A-Za-z]+)`)
	headerRegexp = regexp.MustCompile(`(?:-H|--header)\s+(?:'([^']+)'|"([^"]+)")`)
	cookieRegexp = regexp.MustCompile(`(?:-b|--cookie)\s+(?:'([^']+)'|"([^"]+)")`)
	bodyRegexp   = regexp.MustCompile(`(?:-d|--data|--data-raw|--data-binary)\s+(?:'([^']*)'|"([^"]*)")`)
)

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// ParseCurl 解析 cURL 命令行参数
func ParseCurl(command string) (string, string, map[string]string, string, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return "", "", nil, "", errors.New("curl 命令不能为空")
	}

	var url string
	method := "GET"
	headers := make(map[string]string)
	var body string

	// 提取 URL
	if match := urlRegexp.FindStringSubmatch(command); len(match) > 1 {
		url = firstNonEmpty(match[1], match[2], match[3])
	}

	// 提取请求方法 (-X 或 --request)
	if match := methodRegexp.FindStringSubmatch(command); len(match) > 1 {
		method = strings.ToUpper(match[1])
	}

	// 提取请求头 (-H 或 --header)
	for _, match := range headerRegexp.FindAllStringSubmatch(command, -1) {
		if len(match) > 1 {
			headerStr := firstNonEmpty(match[1], match[2])
			if idx := strings.Index(headerStr, ":"); idx != -1 {
				k := strings.TrimSpace(headerStr[:idx])
				v := strings.TrimSpace(headerStr[idx+1:])
				if k != "" {
					headers[k] = v
				}
			}
		}
	}

	// 提取 Cookie (-b 或 --cookie)
	if match := cookieRegexp.FindStringSubmatch(command); len(match) > 1 {
		cookieVal := firstNonEmpty(match[1], match[2])
		if cookieVal != "" {
			headers["Cookie"] = cookieVal
		}
	}

	// 提取请求体 (-d / --data / --data-raw / --data-binary)
	if match := bodyRegexp.FindStringSubmatch(command); len(match) > 1 {
		body = firstNonEmpty(match[1], match[2])
		if method == "GET" {
			method = "POST"
		}
	}

	if url == "" {
		return "", "", nil, "", errors.New("无法从 curl 命令中提取有效的 URL")
	}

	return url, method, headers, body, nil
}

// BuildCurl 根据 Task 参数重新生成标准 cURL 命令
func BuildCurl(task Task) string {
	parts := []string{fmt.Sprintf("curl '%s'", task.URL)}
	if task.Method != "" && task.Method != "GET" {
		parts = append(parts, fmt.Sprintf("-X %s", task.Method))
	}
	for k, v := range task.Headers {
		parts = append(parts, fmt.Sprintf("-H '%s: %s'", k, v))
	}
	if task.Body != "" {
		escapedBody := strings.ReplaceAll(task.Body, "'", `'"'"'`)
		parts = append(parts, fmt.Sprintf("-d '%s'", escapedBody))
	}
	return strings.Join(parts, " \\\n  ")
}

// NormalizeTask 规范化并校验任务参数
func NormalizeTask(task Task) (Task, error) {
	if strings.TrimSpace(task.CurlCommand) != "" {
		pURL, pMethod, pHeaders, pBody, err := ParseCurl(task.CurlCommand)
		if err == nil {
			if task.URL == "" {
				task.URL = pURL
			}
			if len(task.Headers) == 0 {
				task.Headers = pHeaders
			}
			if task.Body == "" {
				task.Body = pBody
			}
			if task.Method == "" {
				task.Method = pMethod
			}
		}
	}

	task.Name = strings.TrimSpace(task.Name)
	task.URL = strings.TrimSpace(task.URL)
	task.Method = strings.ToUpper(strings.TrimSpace(task.Method))

	if task.Name == "" {
		return task, errors.New("任务名称为必填项")
	}
	if !strings.HasPrefix(task.URL, "http://") && !strings.HasPrefix(task.URL, "https://") {
		return task, errors.New("任务 URL 必须以 http:// 或 https:// 开头")
	}
	if task.Method == "" {
		task.Method = "GET"
	}
	validMethods := map[string]bool{"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true}
	if !validMethods[task.Method] {
		return task, errors.New("不支持的 HTTP 请求方法")
	}

	if task.TimeoutSeconds < 1 {
		task.TimeoutSeconds = 30
	} else if task.TimeoutSeconds > 120 {
		task.TimeoutSeconds = 120
	}

	if task.RetryCount < 0 {
		task.RetryCount = 0
	} else if task.RetryCount > 5 {
		task.RetryCount = 5
	}

	if task.ScheduleHour < 0 || task.ScheduleHour > 23 {
		task.ScheduleHour = 8
	}
	if task.ScheduleMinute < 0 || task.ScheduleMinute > 59 {
		task.ScheduleMinute = 0
	}
	if task.ScheduleSecond < 0 || task.ScheduleSecond > 59 {
		task.ScheduleSecond = 0
	}

	if task.CurlCommand == "" {
		task.CurlCommand = BuildCurl(task)
	}

	return task, nil
}
