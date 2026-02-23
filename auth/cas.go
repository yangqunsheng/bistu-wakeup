package auth

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// LoginParams CAS 登录页面的隐藏参数
type LoginParams struct {
	LT        string
	Execution string
	EventID   string
	RmShown   string
	Salt      string
	ActionURL string
}

// extractSalt 从页面中提取加密 salt（优先 hidden input，fallback JS 变量）
func extractSalt(doc *goquery.Document, html string) string {
	// 优先：<input type="hidden" id="pwdEncryptSalt" value="...">
	if val, exists := doc.Find("#pwdEncryptSalt").Attr("value"); exists && val != "" {
		return val
	}
	// fallback：var pwdDefaultEncryptSalt = "..."
	re := regexp.MustCompile(`pwdDefaultEncryptSalt\s*=\s*"([^"]+)"`)
	if m := re.FindStringSubmatch(html); len(m) > 1 {
		return m[1]
	}
	return ""
}

// extractLoginParams 从 goquery 文档中提取登录参数
func extractLoginParams(doc *goquery.Document) *LoginParams {
	params := &LoginParams{EventID: "submit", RmShown: "1"}

	doc.Find("input[name='lt']").Each(func(_ int, s *goquery.Selection) {
		params.LT, _ = s.Attr("value")
	})
	doc.Find("input[name='execution']").Each(func(_ int, s *goquery.Selection) {
		params.Execution, _ = s.Attr("value")
	})

	// 优先从密码表单 #pwdFromId 获取 action，fallback #casLoginForm
	for _, sel := range []string{"#pwdFromId", "#casLoginForm"} {
		doc.Find(sel).Each(func(_ int, s *goquery.Selection) {
			if params.ActionURL == "" {
				params.ActionURL, _ = s.Attr("action")
			}
		})
	}

	html, _ := doc.Html()
	params.Salt = extractSalt(doc, html)

	return params
}

// parseLoginPage 从 CAS 登录页面提取隐藏字段和加密 salt（仅用于测试）
func parseLoginPage(loginURL string) (*LoginParams, error) {
	resp, err := http.Get(loginURL)
	if err != nil {
		return nil, fmt.Errorf("请求登录页失败: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("解析 HTML 失败: %w", err)
	}

	return extractLoginParams(doc), nil
}

// CASLogin 执行 CAS 统一身份认证登录
// 每次调用使用全新的 cookie jar，避免上次失败的 cookie 污染
func (c *Client) CASLogin(username, password string) error {
	// 关键：每次登录尝试使用干净的 cookie jar
	jar, _ := cookiejar.New(nil)
	c.HTTP.Jar = jar

	loginURL := CASLoginURL + "?service=" + url.QueryEscape(ServiceURL)

	// 1. GET 登录页，提取参数
	resp, err := c.HTTP.Get(loginURL)
	if err != nil {
		return fmt.Errorf("请求登录页失败: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("读取登录页失败: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("解析登录页失败: %w", err)
	}

	params := extractLoginParams(doc)

	// 2. 检查 salt
	if params.Salt == "" {
		// 诊断：页面可能不是正常登录页（验证码页、锁定页等）
		title := doc.Find("title").Text()
		errText := ""
		doc.Find("#showErrorTip span, .auth_error, #errorMsg").Each(func(_ int, s *goquery.Selection) {
			if t := strings.TrimSpace(s.Text()); t != "" {
				errText = t
			}
		})
		if errText != "" {
			return fmt.Errorf("CAS 服务器提示: %s", errText)
		}
		return fmt.Errorf("未获取到加密 salt（页面标题: %q），登录页面结构可能已变化", title)
	}

	// 3. 加密密码
	encryptedPwd, err := EncryptPassword(password, params.Salt)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 4. POST 登录
	formData := url.Values{
		"username":  {username},
		"password":  {encryptedPwd},
		"lt":        {params.LT},
		"execution": {params.Execution},
		"_eventId":  {params.EventID},
		"rmShown":   {params.RmShown},
		"dllt":      {"generalLogin"},
		"cllt":      {"userNameLogin"},
	}

	resp, err = c.HTTP.PostForm(loginURL, formData)
	if err != nil {
		return fmt.Errorf("登录请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 5. 判断结果
	finalURL := resp.Request.URL.String()
	if !strings.Contains(finalURL, "authserver/login") {
		return nil // 成功：已跳转离开登录页
	}

	// 登录失败，提取具体错误信息
	respBody, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("登录失败: %s", extractErrorMsg(string(respBody)))
}

// extractErrorMsg 从 CAS 响应页面提取错误信息
func extractErrorMsg(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "未知错误"
	}

	// 优先提取 #showErrorTip 中的错误信息
	var msg string
	doc.Find("#showErrorTip span").Each(func(_ int, s *goquery.Selection) {
		if t := strings.TrimSpace(s.Text()); t != "" {
			msg = t
		}
	})

	// 过滤掉无意义的错误信息（如"图形动态码"等）
	if msg != "" {
		// 常见的真实错误信息
		validErrors := []string{
			"用户名或者密码有误",
			"用户名或密码有误",
			"账号已锁定",
			"账号不存在",
			"密码错误",
		}
		for _, valid := range validErrors {
			if strings.Contains(msg, valid) {
				return msg
			}
		}
		// 如果包含"验证码"但实际上不需要验证码，忽略此错误
		if strings.Contains(msg, "验证码") || strings.Contains(msg, "动态码") {
			return "用户名或密码错误"
		}
		return msg
	}

	return "用户名或密码错误"
}

// NeedCaptcha 检查是否需要验证码
func (c *Client) NeedCaptcha(username string) (bool, error) {
	resp, err := c.HTTP.Get(NeedCaptchaURL + "?username=" + url.QueryEscape(username))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	buf := make([]byte, 64)
	n, _ := resp.Body.Read(buf)
	return strings.TrimSpace(string(buf[:n])) == "true", nil
}
