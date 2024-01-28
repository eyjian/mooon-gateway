// Package middleware
// Written by yijian on 2024/01/28
package middleware

import (
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"io"
	"mooon-gateway/mooonlogin"
	"net/http"
	"strings"
)
import (
	"mooon-gateway/pb/mooon_login"
)

// LoginMiddleware 登录
func LoginMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logCtx := logx.ContextWithFields(r.Context(), logx.Field("path", r.URL.Path))

		if !strings.HasPrefix(r.URL.Path, "/v1/") {
			next.ServeHTTP(w, r)
		} else {
			var loginReq mooon_login.LoginReq
			var mooonLogin mooonlogin.MooonLogin

			reqBodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				logc.Errorf(logCtx, "Read request body error: %s\n", err.Error())
				return
			}
			defer r.Body.Close()

			loginReq.Body = reqBodyBytes
			loginResp, err := mooonLogin.Login(r.Context(), &loginReq)
			if err != nil {
				logc.Errorf(logCtx, "Call login failed: %s\n", err.Error())
				responseBytes, err := NewResponseStr(logCtx, GwErrCallLogin, "call login error", nil)
				if err == nil {
					w.Header().Set("Content-Type", "application/json")
					w.Write(responseBytes)
					return
				}
			}

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
			responseBytes, err := NewResponseStr(logCtx, GwSuccess, "", loginResp.Body)
			if err == nil {
				_, err = w.Write(responseBytes) // 得放在最后
				if err != nil {
					logc.Errorf(logCtx, "Write response: %s\n", err.Error())
				} else {
					w.WriteHeader(http.StatusOK)
				}
			}
		}
	}
}
