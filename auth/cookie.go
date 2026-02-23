package auth

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// CookieLogin 使用用户提供的 Cookie 字符串设置认证
func (c *Client) CookieLogin(baseURL, cookieStr string) error {
	u, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("解析 URL 失败: %w", err)
	}

	var cookies []*http.Cookie
	for _, part := range strings.Split(cookieStr, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		cookies = append(cookies, &http.Cookie{
			Name:  strings.TrimSpace(kv[0]),
			Value: strings.TrimSpace(kv[1]),
		})
	}

	if len(cookies) == 0 {
		return fmt.Errorf("未解析到有效的 Cookie")
	}

	c.HTTP.Jar.SetCookies(u, cookies)
	return nil
}
