package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"

	"github.com/bistu-wakeup/bistu-wakeup/auth"
	"github.com/bistu-wakeup/bistu-wakeup/export"
	"github.com/bistu-wakeup/bistu-wakeup/schedule"
)

const version = "0.2.0"

// 颜色定义
var (
	cyan    = color.New(color.FgCyan, color.Bold).SprintFunc()
	green   = color.New(color.FgGreen, color.Bold).SprintFunc()
	yellow  = color.New(color.FgYellow, color.Bold).SprintFunc()
	magenta = color.New(color.FgMagenta, color.Bold).SprintFunc()
	blue    = color.New(color.FgBlue).SprintFunc()
	dim     = color.New(color.Faint).SprintFunc()
	bold    = color.New(color.Bold).SprintFunc()
)

func main() {
	printBanner()

	if err := run(); err != nil {
		fmt.Printf("\n  %s %s\n\n", color.RedString("✗"), err)
		os.Exit(1)
	}
}

func printBanner() {
	fmt.Println()
	fmt.Printf("  %s\n", cyan("╔═══════════════════════════════════════════╗"))
	fmt.Printf("  %s\n", cyan("║")+"  "+bold("BISTU 课表导出工具")+"                  "+cyan("║"))
	fmt.Printf("  %s\n", cyan("║")+"  "+dim("WakeUp 格式 · v"+version)+"                 "+cyan("║"))
	fmt.Printf("  %s\n", cyan("╚═══════════════════════════════════════════╝"))
	fmt.Println()
}

func run() error {
	cookieStr := flag.String("cookie", "", "使用 Cookie 模式（高级用户）")
	flag.Parse()

	// 1. 认证
	client, err := auth.NewClient()
	if err != nil {
		return fmt.Errorf("初始化失败: %w", err)
	}

	printStep(1, 4, "身份认证")
	if *cookieStr != "" {
		fmt.Printf("    %s 使用 Cookie 模式\n", blue("→"))
		if err := client.CookieLogin("https://jwxt.bistu.edu.cn", *cookieStr); err != nil {
			return err
		}
		fmt.Printf("    %s Cookie 已设置\n\n", green("✓"))
	} else {
		if err := interactiveLogin(client); err != nil {
			return err
		}
	}

	// 2. 获取用户信息
	printStep(2, 4, "获取用户信息")
	fetcher := &schedule.Fetcher{Client: client.HTTP}
	userInfo, err := fetcher.FetchUserInfo()
	if err != nil {
		return err
	}
	welcome := userInfo.StudentID
	if userInfo.UserName != "" {
		welcome = fmt.Sprintf("%s (%s)", userInfo.UserName, userInfo.StudentID)
	}
	fmt.Printf("    %s 欢迎, %s\n\n", green("✓"), bold(welcome))

	// 3. 选择学期
	printStep(3, 4, "选择学期")
	termCode, err := selectTerm(userInfo)
	if err != nil {
		return err
	}

	// 4. 获取课表
	printStep(4, 4, "获取课表")
	rawCourses, err := fetcher.FetchSchedule(termCode, userInfo.StudentID)
	if err != nil {
		return err
	}
	fmt.Printf("    %s 获取到 %s 门课程\n\n", green("✓"), bold(fmt.Sprintf("%d", len(rawCourses))))

	// 5. 解析并导出
	courses := schedule.ParseAll(rawCourses)
	rows := make([][]string, 0, len(courses))
	for _, c := range courses {
		rows = append(rows, []string{
			c.Name, c.DayOfWeek, c.BeginSection,
			c.EndSection, c.Teacher, c.Location, c.Weeks,
		})
	}

	filename := fmt.Sprintf("schedule_%s.csv", termCode)
	if err := export.WriteCSV(filename, rows); err != nil {
		return err
	}

	// 完成
	fmt.Println(cyan("  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	fmt.Printf("\n  %s %s\n", green("✓"), bold("导出成功!"))
	fmt.Printf("    %s %s\n", magenta("📄"), bold("./"+filename))
	fmt.Printf("    %s %d 门课程\n\n", blue("📊"), len(rawCourses))
	fmt.Printf("  %s\n", dim("💡 提示: 打开 WakeUp → 导入课表 → 选择此文件"))
	fmt.Println()
	return nil
}

func printStep(current, total int, title string) {
	bar := ""
	for i := 1; i <= total; i++ {
		if i < current {
			bar += green("●")
		} else if i == current {
			bar += cyan("●")
		} else {
			bar += dim("○")
		}
		if i < total {
			bar += dim("─")
		}
	}
	fmt.Printf("  %s %s\n\n", bar, bold(title))
}

func interactiveLogin(client *auth.Client) error {
	// Windows 下禁用 promptui 的 ANSI 渲染，避免重复打印
	usernamePrompt := promptui.Prompt{
		Label:  "学号",
		Stdout: &bellSkipper{},
	}
	username, err := usernamePrompt.Run()
	if err != nil {
		return fmt.Errorf("输入取消")
	}

	needCaptcha, _ := client.NeedCaptcha(username)
	if needCaptcha {
		fmt.Printf("\n    %s 当前需要验证码（短时间内尝试过多）\n", yellow("⚠"))
		sel := promptui.Select{
			Label: "请选择",
			Items: []string{
				"等待 30 秒后重试",
				"切换到 Cookie 模式",
				"退出",
			},
		}
		idx, _, _ := sel.Run()
		switch idx {
		case 0:
			fmt.Printf("    %s 等待 30 秒...\n", blue("⏳"))
			time.Sleep(30 * time.Second)
			return interactiveLogin(client)
		case 1:
			fmt.Printf("\n  请从浏览器开发者工具复制 Cookie，然后运行:\n")
			fmt.Printf("  %s\n\n", bold(`bistu-wakeup --cookie "JSESSIONID=xxx; route=xxx"`))
			os.Exit(0)
		default:
			os.Exit(0)
		}
	}

	for attempt := 0; attempt < 3; attempt++ {
		pwdPrompt := promptui.Prompt{
			Label:  "密码",
			Mask:   '*',
			Stdout: &bellSkipper{},
		}
		password, err := pwdPrompt.Run()
		if err != nil {
			return fmt.Errorf("输入取消")
		}

		fmt.Printf("    %s 正在登录...\n", blue("→"))
		err = client.CASLogin(username, password)
		if err == nil {
			fmt.Printf("    %s 登录成功\n\n", green("✓"))
			return nil
		}

		fmt.Printf("    %s %v\n", color.RedString("✗"), err)
		if attempt < 2 {
			retryPrompt := promptui.Prompt{
				Label:     "重新输入密码",
				IsConfirm: true,
				Stdout:    &bellSkipper{},
			}
			if _, err := retryPrompt.Run(); err != nil {
				return fmt.Errorf("登录取消")
			}
		}
	}
	return fmt.Errorf("登录失败次数过多，请稍后再试或使用 --cookie 模式")
}

// bellSkipper 实现 io.WriteCloser，过滤掉 promptui 的 bell 字符和 ANSI 控制码
type bellSkipper struct{}

func (bs *bellSkipper) Write(b []byte) (int, error) {
	const charBell = 7 // bell 字符
	if len(b) == 1 && b[0] == charBell {
		return 0, nil
	}
	return os.Stderr.Write(b)
}

func (bs *bellSkipper) Close() error {
	return os.Stderr.Close()
}

func selectTerm(info *schedule.UserInfo) (string, error) {
	// 生成最近 8 个学期（包含小学期）
	terms := schedule.GenerateRecentTerms(time.Now(), 8)

	// 构建选项列表
	items := make([]string, 0, len(terms)+1)
	for _, t := range terms {
		prefix := "  "
		if t.IsCurrent {
			prefix = green("★ ")
		}
		items = append(items, prefix+t.Label)
	}
	items = append(items, dim("  ✏  手动输入学期代码..."))

	sel := promptui.Select{
		Label: "    " + dim("请选择学期"),
		Items: items,
		Size:  len(items),
		Templates: &promptui.SelectTemplates{
			Active:   fmt.Sprintf("%s {{ . | cyan }}", cyan("▶")),
			Inactive: "  {{ . }}",
			Selected: fmt.Sprintf("    %s {{ . }}", green("✓")),
		},
	}

	idx, _, err := sel.Run()
	if err != nil {
		return "", fmt.Errorf("选择取消")
	}

	// 手动输入
	if idx == len(terms) {
		fmt.Println()
		fmt.Printf("    %s\n", dim("格式说明:"))
		fmt.Printf("      %s  第一学期 (秋季)\n", cyan("YYYY-YYYY-1"))
		fmt.Printf("      %s  第二学期 (春季)\n", cyan("YYYY-YYYY-2"))
		fmt.Printf("      %s  小学期 (夏季)\n", cyan("YYYY-YYYY-3"))
		fmt.Println()

		codePrompt := promptui.Prompt{
			Label:   "学期代码",
			Default: "2025-2026-2",
			Stdout:  &bellSkipper{},
		}
		code, err := codePrompt.Run()
		if err != nil {
			return "", fmt.Errorf("输入取消")
		}
		code = strings.TrimSpace(code)
		if code == "" {
			return "", fmt.Errorf("学期代码不能为空")
		}
		fmt.Println()
		return code, nil
	}

	fmt.Println()
	return terms[idx].Code, nil
}
