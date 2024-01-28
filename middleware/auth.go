// Package middleware
// Written by yijian on 2024/01/28
package middleware

import (
	"io"
	"net/http"
	"strings"
)
import (
	"mooon-gateway/mooonauth"
	"mooon-gateway/pb/mooon_auth"
)

// AuthMiddleware 鉴权
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v2/") {
			next.ServeHTTP(w, r)
		} else {
			var authReq mooon_auth.AuthReq
			var mooonAuth mooonauth.MooonAuth

			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				return
			} else {
				defer r.Body.Close()

				authReq.Body = bodyBytes
				authResp, err := mooonAuth.Authenticate(r.Context(), &authReq)
				if err == nil {
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
					if len(authResp.Body) > 0 {
						_, _ = w.Write(authResp.Body) // 得放在最后
					}
					w.WriteHeader(http.StatusOK)
				} else {
					return
				}
			}
		}
	}
}
