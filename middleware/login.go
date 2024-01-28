// Package middleware
// Written by yijian on 2024/01/28
package middleware

import (
	"io"
	"net/http"
	"strings"
)
import (
	"mooon-gateway/mooonlogin"
	"mooon-gateway/pb/mooon_login"
)

// LoginMiddleware 登录
func LoginMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v1/") {
			next.ServeHTTP(w, r)
		} else {
			var loginReq mooon_login.LoginReq
			var mooonLogin mooonlogin.MooonLogin

			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				return
			} else {
				defer r.Body.Close()

				loginReq.Body = bodyBytes
				loginResp, err := mooonLogin.Login(r.Context(), &loginReq)
				if err == nil {
					// 写 http 头
					for name, value := range loginResp.HttpHeaders {
						w.Header().Set(name, value)
					}
					// 写 cookies
					for _, loginCookie := range loginResp.HttpCookies {
						httpCookie := LoginCookie2HttpCookie(loginCookie)
						http.SetCookie(w, httpCookie)
					}
					// 写响应体
					if len(loginResp.Body) > 0 {
						_, _ = w.Write(loginResp.Body) // 得放在最后
					}
					w.WriteHeader(http.StatusOK)
				} else {
					return
				}
			}
		}
	}
}
