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
    "github.com/zeromicro/go-zero/core/conf"
    "github.com/zeromicro/go-zero/core/logc"
    "github.com/zeromicro/go-zero/core/logx"
    "github.com/zeromicro/go-zero/zrpc"
)
import (
    "mooon-gateway/mooonlogin"
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

            mooonLogin, err := getLoginClient(logCtx)
            if err != nil {
                responseBytes, err := NewResponseStr(logCtx, GwErrConnLogin, "conn login error", nil)
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

            loginReq.Body = reqBodyBytes
            loginResp, err := mooonLogin.Login(r.Context(), &loginReq)
            if err != nil {
                logc.Errorf(logCtx, "Call login failed: %s\n", err.Error())
                responseBytes, err := NewResponseStr(logCtx, GwErrCallLogin, "call login error", nil)
                if err == nil {
                    w.Header().Set("Content-Type", "application/json")
                    w.Write(responseBytes)
                }
                return
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
            if len(loginResp.Body) > 0 {
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
}

func getLoginClient(logCtx context.Context) (mooonlogin.MooonLogin, error) {
    var loginConf zrpc.RpcClientConf

    err := conf.Load("etc/login.yaml", &loginConf)
    if err != nil {
        logc.Errorf(logCtx, "Load conf error: %s\n", err.Error())
        return nil, err
    }

    zrpcClient, err := zrpc.NewClient(loginConf)
    if err != nil {
        logc.Errorf(logCtx, "New login client error: %s\n", err.Error())
        return nil, err
    }

    loginClient := mooonlogin.NewMooonLogin(zrpcClient)
    return loginClient, nil
}