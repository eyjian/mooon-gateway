// Package middleware
// Written by yijian on 2024/01/28
package middleware

import (
    "context"
    "google.golang.org/grpc/status"
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
    "mooon-gateway/mooonlogin"
    "mooon-gateway/pb/mooon_login"
)

// LoginMiddleware 登录
func LoginMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        logCtx := logx.ContextWithFields(r.Context(),
            logx.Field("method", r.Method),
            logx.Field("host", r.Host),
            logx.Field("path", r.URL.Path),
            logx.Field("remote", r.RemoteAddr))

        logc.Debugf(logCtx, "Login.Prefix: %s, r.URL.Path: %s\n", GlobalConfig.Login.Prefix, r.URL.Path)
        if !strings.HasPrefix(r.URL.Path, GlobalConfig.Login.Prefix) {
            next.ServeHTTP(w, r)
        } else {
            loginHandle(logCtx, w, r)
        }
    }
}

func getLoginClient(logCtx context.Context) (mooonlogin.MooonLogin, error) {
    /*var loginConf zrpc.RpcClientConf

      err := conf.Load("etc/login.yaml", &loginConf)
      if err != nil {
      	logc.Errorf(logCtx, "Load conf error: %s\n", err.Error())
      	return nil, err
      }*/

    zrpcClient, err := zrpc.NewClient(GlobalConfig.Login.RpcClientConf)
    if err != nil {
        logc.Errorf(logCtx, "New login client error: %s\n", err.Error())
        return nil, err
    }

    loginClient := mooonlogin.NewMooonLogin(zrpcClient)
    return loginClient, nil
}

func loginHandle(logCtx context.Context, w http.ResponseWriter, r *http.Request) {
    var loginReq mooon_login.LoginReq

    // 实例化登录服务
    mooonLogin, err := getLoginClient(logCtx)
    if err != nil {
        responseBytes, err := NewResponseStr(logCtx, GwErrConnLogin, "connect login error", "")
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

    loginReq.Body = string(reqBodyBytes)
    loginResp, err := mooonLogin.Login(r.Context(), &loginReq)
    if err != nil {
        // 登录失败或者出错
        var code ErrCode
        var message string

        // 处理出错码
        if st, ok := status.FromError(err); ok {
            code = ErrCode(st.Code())
            message = st.Message()
            logc.Errorf(logCtx, "call login error: (%d) %s", code, message)
        } else {
            code = GwErrCallLogin
            message = "call login failed"
            logc.Errorf(logCtx, "%s: %s\n", message, err.Error())
        }

        // 出错响应
        responseBytes, _ := NewResponseStr(logCtx, code, message, "")
        w.Header().Set("Content-Type", "application/json")
        w.Write(responseBytes)

        return
    } else { // 登录成功
        // 写 http 头
        w.Header().Set("Content-Type", "application/json")
        for name, value := range loginResp.HttpHeaders {
            w.Header().Set(name, value)
        }
        // 写 cookies
        for _, loginCookie := range loginResp.HttpCookies {
            httpCookie := LoginCookie2HttpCookie(loginCookie)
            http.SetCookie(w, httpCookie)
        }

        // 写响应体
        responseBytes, err := NewResponseStr(logCtx, GwSuccess, "success", loginResp.Body)
        if err != nil {
            logc.Errorf(logCtx, "marshal response error: %s (%s)\n", err.Error(), loginResp.Body)
            responseBytes, _ := NewResponseStr(logCtx, GwInvalidResp, "marshal response error", "")
            w.Write(responseBytes)
        } else {
            _, err = w.Write(responseBytes) // 得放在最后
            if err != nil {
                logc.Errorf(logCtx, "write response error: %s\n", err.Error())
            } else {
                w.WriteHeader(http.StatusOK)
                logc.Infof(logCtx, "login success: host=%s, path=%s, remote=%s", r.Host, r.URL.Path, r.RemoteAddr)
            }
        }
    }
}
