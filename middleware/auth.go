// Package middleware
// Written by yijian on 2024/01/28
package middleware

import (
    "context"
    "io"
    "net/http"
    "strings"
)
import (
    "github.com/zeromicro/go-zero/core/logc"
    "github.com/zeromicro/go-zero/core/logx"
    "github.com/zeromicro/go-zero/zrpc"
)
import (
    "mooon-gateway/mooonauth"
    "mooon-gateway/pb/mooon_auth"
)

// AuthMiddleware 鉴权
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        logCtx := logx.ContextWithFields(r.Context(), logx.Field("path", r.URL.Path))

        logc.Debugf(logCtx, "Auth.Prefix: %s, r.URL.Path: %s\n", GlobalConfig.Auth.Prefix, r.URL.Path)
        if !strings.HasPrefix(r.URL.Path, GlobalConfig.Auth.Prefix) {
            next.ServeHTTP(w, r)
        } else {
            var authReq mooon_auth.AuthReq

            mooonAuth, err := getAuthClient(logCtx)
            if err != nil {
                responseBytes, err := NewResponseStr(logCtx, GwErrConnAuth, "conn auth error", nil)
                if err == nil {
                    w.Header().Set("Content-Type", "application/json")
                    w.Write(responseBytes)
                }
                return
            }

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
            if len(authResp.Body) > 0 {
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
}

func getAuthClient(logCtx context.Context) (mooonauth.MooonAuth, error) {
    /*var authConf zrpc.RpcClientConf

      err := conf.Load("etc/auth.yaml", &authConf)
      if err != nil {
      	logc.Errorf(logCtx, "Load conf error: %s\n", err.Error())
      	return nil, err
      }*/

    zrpcClient, err := zrpc.NewClient(GlobalConfig.Auth.RpcClientConf)
    if err != nil {
        logc.Errorf(logCtx, "New auth client error: %s\n", err.Error())
        return nil, err
    }

    authClient := mooonauth.NewMooonAuth(zrpcClient)
    return authClient, nil
}
