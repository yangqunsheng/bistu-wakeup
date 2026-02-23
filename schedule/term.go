package schedule

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Term 学期信息
type Term struct {
	Code      string
	IsCurrent bool
	Label     string
}

// SortTerms 根据当前日期智能排序学期列表
func SortTerms(codes []string, now time.Time) []Term {
	currentCode := guessCurrentTerm(now)

	terms := make([]Term, 0, len(codes))
	for _, code := range codes {
		isCurrent := code == currentCode
		label := code
		if isCurrent {
			label = "★ " + code + " (当前学期)"
		} else if strings.HasSuffix(code, "-3") {
			label = code + " (小学期)"
		}
		terms = append(terms, Term{Code: code, IsCurrent: isCurrent, Label: label})
	}

	sort.SliceStable(terms, func(i, j int) bool {
		if terms[i].IsCurrent {
			return true
		}
		if terms[j].IsCurrent {
			return false
		}
		return termWeight(terms[i].Code) > termWeight(terms[j].Code)
	})

	return terms
}

func guessCurrentTerm(now time.Time) string {
	year := now.Year()
	month := int(now.Month())

	var startYear, endYear, n int
	switch {
	case month >= 9:
		startYear, endYear, n = year, year+1, 1
	case month <= 1:
		startYear, endYear, n = year-1, year, 1
	case month >= 2 && month <= 6:
		startYear, endYear, n = year-1, year, 2
	default:
		startYear, endYear, n = year-1, year, 3
	}

	return fmt.Sprintf("%d-%d-%d", startYear, endYear, n)
}

func termWeight(code string) int {
	parts := strings.Split(code, "-")
	if len(parts) != 3 {
		return 0
	}
	startYear, _ := strconv.Atoi(parts[0])
	n, _ := strconv.Atoi(parts[2])
	return startYear*10 + n
}

// GenerateRecentTerms 生成学期列表（从未来到过去，包含小学期）
// 按 BISTU 实际顺序：2026-2027-1, 2025-2026-3, 2025-2026-2, 2025-2026-1, ...
func GenerateRecentTerms(now time.Time, count int) []Term {
	current := guessCurrentTerm(now)

	// 从当前学期向未来推 1 年，再向过去推
	parts := strings.Split(current, "-")
	if len(parts) != 3 {
		return []Term{{Code: current, IsCurrent: true, Label: FormatTermLabel(current, true)}}
	}

	startYear, _ := strconv.Atoi(parts[0])
	semester, _ := strconv.Atoi(parts[2])

	// 向未来推 1 个学年（确保包含未来学期）
	startYear++
	semester = 1

	terms := make([]Term, 0, count)
	for i := 0; i < count; i++ {
		code := fmt.Sprintf("%d-%d-%d", startYear, startYear+1, semester)
		isCurrent := code == current
		terms = append(terms, Term{
			Code:      code,
			IsCurrent: isCurrent,
			Label:     FormatTermLabel(code, isCurrent),
		})

		// 向过去推：1 → 3(上一学年) → 2 → 1(再上一学年)
		if semester == 1 {
			startYear-- // 回到上一学年
			semester = 3
		} else if semester == 3 {
			semester = 2 // 同一学年
		} else { // semester == 2
			semester = 1 // 同一学年
		}
	}
	return terms
}

// FormatTermLabel 格式化学期显示标签
func FormatTermLabel(code string, isCurrent bool) string {
	parts := strings.Split(code, "-")
	if len(parts) != 3 {
		return code
	}
	name := "第一学期"
	switch parts[2] {
	case "2":
		name = "第二学期"
	case "3":
		name = "小学期"
	}
	label := fmt.Sprintf("%s-%s学年 %s", parts[0], parts[1], name)
	if isCurrent {
		label += "  (当前)"
	}
	return label
}
