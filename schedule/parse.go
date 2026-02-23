package schedule

import (
	"fmt"
	"regexp"
	"strings"
)

// Course 结构化课程数据
type Course struct {
	Name         string
	DayOfWeek    string
	BeginSection string
	EndSection   string
	Teacher      string
	Location     string
	Weeks        string
}

var bracketRe = regexp.MustCompile(`\[.*?\]`)

// ParseCourse 从原始 API 数据解析为结构化课程
func ParseCourse(raw map[string]interface{}) Course {
	c := Course{
		Name:         getStr(raw, "courseName", "无"),
		DayOfWeek:    getStr(raw, "dayOfWeek", "无"),
		BeginSection: getStr(raw, "beginSection", "无"),
		EndSection:   getStr(raw, "endSection", "无"),
		Location:     getStr(raw, "placeName", "无"),
	}

	wt := getStr(raw, "weeksAndTeachers", "")
	if wt != "" {
		parts := strings.SplitN(wt, "/", 2)
		if len(parts) > 0 {
			weeks := parts[0]
			weeks = bracketRe.ReplaceAllString(weeks, "")
			weeks = strings.ReplaceAll(weeks, "周", "")
			weeks = strings.TrimSpace(weeks)
			if weeks != "" {
				c.Weeks = weeks
			}
		}
		if len(parts) > 1 {
			teacher := parts[1]
			teacher = bracketRe.ReplaceAllString(teacher, "")
			teacher = strings.TrimSpace(teacher)
			if teacher != "" {
				c.Teacher = teacher
			}
		}
	}

	if c.Teacher == "" {
		c.Teacher = "无"
	}
	if c.Location == "" {
		c.Location = "无"
	}
	if c.Weeks == "" {
		c.Weeks = "无"
	}

	return c
}

// ParseAll 批量解析课程列表
func ParseAll(rawList []map[string]interface{}) []Course {
	courses := make([]Course, 0, len(rawList))
	for _, raw := range rawList {
		courses = append(courses, ParseCourse(raw))
	}
	return courses
}

func getStr(m map[string]interface{}, key, fallback string) string {
	if v, ok := m[key]; ok {
		s := fmt.Sprintf("%v", v)
		s = strings.TrimSpace(s)
		if s != "" && s != "<nil>" {
			return s
		}
	}
	return fallback
}
