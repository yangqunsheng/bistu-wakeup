package export

import (
	"fmt"
	"os"
	"strings"
)

var header = []string{"课程名称", "星期", "开始节数", "结束节数", "老师", "地点", "周数"}

// WriteCSV 生成 WakeUp 格式的 CSV 文件
func WriteCSV(filename string, courses [][]string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer f.Close()

	// UTF-8 BOM
	f.WriteString("\uFEFF")

	// 表头
	f.WriteString(formatRow(header) + "\n")

	// 数据行
	for _, row := range courses {
		f.WriteString(formatRow(row) + "\n")
	}

	return nil
}

// formatRow 将一行数据格式化为 CSV 行（双引号包裹，逗号分隔）
func formatRow(fields []string) string {
	quoted := make([]string, len(fields))
	for i, f := range fields {
		escaped := strings.ReplaceAll(f, `"`, `""`)
		quoted[i] = `"` + escaped + `"`
	}
	return strings.Join(quoted, ",")
}
