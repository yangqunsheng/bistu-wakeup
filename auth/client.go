package auth

import (
	"net/http"
	"net/http/cookiejar"
)

const (
	CASLoginURL    = "https://wxjw.bistu.edu.cn/authserver/login"
	ServiceURL     = "https://jwxt.bistu.edu.cn/jwapp/sys/homeapp/index.do"
	NeedCaptchaURL = "https://wxjw.bistu.edu.cn/authserver/needCaptcha.html"
)

// Client 封装带 Cookie 管理的 HTTP 客户端
type Client struct {
	HTTP *http.Client
}

// NewClient 创建新的认证客户端
func NewClient() (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &Client{
		HTTP: &http.Client{Jar: jar},
	}, nil
}
