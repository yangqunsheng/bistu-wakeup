package schedule

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	BaseURL        = "https://jwxt.bistu.edu.cn"
	CurrentUserURL = BaseURL + "/jwapp/sys/homeapp/api/home/currentUser.do"
	ScheduleURL    = BaseURL + "/jwapp/sys/homeapp/api/home/student/getMyScheduleDetail.do"
)

// Fetcher 课表数据获取器
type Fetcher struct {
	Client *http.Client
}

// UserInfo 用户信息
type UserInfo struct {
	StudentID string
	UserName  string
	TermCode  string
	Terms     []string
}

// FetchUserInfo 获取当前用户信息和可用学期列表
func (f *Fetcher) FetchUserInfo() (*UserInfo, error) {
	resp, err := f.Client.Get(CurrentUserURL)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	info := &UserInfo{}
	if datas, ok := result["datas"].(map[string]interface{}); ok {
		// 学号优先从 userId 获取
		info.StudentID, _ = datas["userId"].(string)
		info.UserName, _ = datas["userName"].(string)
		if welcome, ok := datas["welcomeInfo"].(map[string]interface{}); ok {
			info.TermCode, _ = welcome["xnxqdm"].(string)
			if info.StudentID == "" {
				info.StudentID, _ = welcome["xh"].(string)
			}
		}
		if info.StudentID == "" {
			if user, ok := datas["user"].(map[string]interface{}); ok {
				info.StudentID, _ = user["xh"].(string)
				if info.StudentID == "" {
					info.StudentID, _ = user["usrId"].(string)
				}
			}
		}
	}

	return info, nil
}

// FetchSchedule 获取指定学期的课表数据
func (f *Fetcher) FetchSchedule(termCode, studentID string) ([]map[string]interface{}, error) {
	formData := url.Values{
		"termCode":    {termCode},
		"studentCode": {studentID},
		"xh":          {studentID},
		"type":        {"term"},
	}

	resp, err := f.Client.Post(ScheduleURL,
		"application/x-www-form-urlencoded;charset=UTF-8",
		strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("获取课表失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	var list []interface{}
	if datas, ok := result["datas"].(map[string]interface{}); ok {
		if l, ok := datas["arrangedList"].([]interface{}); ok {
			list = l
		} else if l, ok := datas["list"].([]interface{}); ok {
			list = l
		}
	}
	if list == nil {
		if data, ok := result["data"].(map[string]interface{}); ok {
			if l, ok := data["rows"].([]interface{}); ok {
				list = l
			}
		}
	}

	if len(list) == 0 {
		return nil, fmt.Errorf("未获取到课程数据，请检查学期代码和学号")
	}

	items := make([]map[string]interface{}, 0, len(list))
	for _, item := range list {
		if m, ok := item.(map[string]interface{}); ok {
			items = append(items, m)
		}
	}
	return items, nil
}
