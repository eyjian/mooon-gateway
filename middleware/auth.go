// Package middleware
// Written by yijian on 2024/01/28
package middleware

import (
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"io"
	"mooon-gateway/mooonauth"
	"mooon-gateway/pb/mooon_auth"
	"net/http"
	"strings"
)

// AuthMiddleware 鉴权
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logCtx := logx.ContextWithFields(r.Context(), logx.Field("path", r.URL.Path))

		if !strings.HasPrefix(r.URL.Path, "/v1/") {
			next.ServeHTTP(w, r)
		} else {
			var authReq mooon_auth.AuthReq
			var mooonAuth mooonauth.MooonAuth

			reqBodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				logc.Errorf(logCtx, "Read request body error: %s\n", err.Error())
				return
			}
			defer r.Body.Close()

			authReq.Body = reqBodyBytes
			authResp, err := mooonAuth.Authenticate(r.Context(), &authReq)
			if err != nil {
				logc.Errorf(logCtx, "Call auth failed: %s\n", err.Error())
				responseBytes, err := NewResponseStr(logCtx, GwErrCallAuth, "call auth error", nil)
				if err == nil {
					w.Write(responseBytes)
					return
				}
			}

			// 写 http 头
			for name, value := range authResp.HttpHeaders {
				w.Header().Set(name, value)
			}
			// 写 cookies
			for _, authCookie := range authResp.HttpCookies {
				httpCookie := AuthCookie2HttpCookie(authCookie)
				http.SetCookie(w, httpCookie)
			}
			// 写响应体
			responseBytes, err := NewResponseStr(logCtx, GwSuccess, "", authResp.Body)
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
