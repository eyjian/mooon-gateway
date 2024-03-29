// Package middleware
// Written by yijian on 2024/01/28
package middleware

import (
    "context"
    "google.golang.org/grpc/status"
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
        logCtx := logx.ContextWithFields(r.Context(),
            logx.Field("method", r.Method),
            logx.Field("host", r.Host),
            logx.Field("path", r.URL.Path),
            logx.Field("remote", r.RemoteAddr))

        logc.Infof(logCtx, "auth request: host=%s, path=%s, remote=%s", r.Host, r.URL.Path, r.RemoteAddr)
        if !strings.HasPrefix(r.URL.Path, GlobalConfig.Auth.Prefix) {
            next.ServeHTTP(w, r) // 不需要鉴权的请求，往下转发
        } else {
            authHandle(next, logCtx, w, r)
        }
    }
}

func getAuthClient(logCtx context.Context) (mooonauth.MooonAuth, error) {
    zrpcClient, err := zrpc.NewClient(GlobalConfig.Auth.RpcClientConf)
    if err != nil {
        logc.Errorf(logCtx, "New auth client error: %s\n", err.Error())
        return nil, err
    }

    authClient := mooonauth.NewMooonAuth(zrpcClient)
    return authClient, nil
}

func authHandle(next http.HandlerFunc, logCtx context.Context, w http.ResponseWriter, r *http.Request) {
    // 实例化鉴权服务
    mooonAuth, err := getAuthClient(logCtx)
    if err != nil {
        responseBytes, err := NewResponseStr(logCtx, GwErrConnAuth, "conn auth error", "")
        if err == nil {
            w.Header().Set("Content-Type", "application/json")
            w.Write(responseBytes)
        }
    } else {
        var authReq mooon_auth.AuthReq

        // http 头（含了 cookies）
        if len(r.Header) > 0 {
            authReq.HttpHeaders = make(map[string]string)
            for key, values := range r.Header {
                for _, value := range values { // 实际不应出现这种重复 key 的情况
                    authReq.HttpHeaders[key] = value
                }
            }
        }

        // http cookies
        httpCookies := r.Cookies()
        if len(httpCookies) > 0 {
            for _, httpCookie := range httpCookies {
                authCookie := HttpCookie2AuthCookie(httpCookie)
                authReq.HttpCookies = append(authReq.HttpCookies, authCookie)
            }
        }

        // 调用鉴权服务
        authResp, err := mooonAuth.Authenticate(r.Context(), &authReq)
        if err != nil { // 鉴权失败或者未通过
            authHandleCallFailure(logCtx, w, err)
        } else { // 鉴权通过，改写请求以加入（传递）鉴权数据
            authHandleCallSuccess(next, logCtx, w, r, authResp)
        }
    }
}

func authHandleCallFailure(logCtx context.Context, w http.ResponseWriter, err error) {
    var code ErrCode
    var message string

    // 处理出错码
    if st, ok := status.FromError(err); ok {
        code = ErrCode(st.Code())
        message = st.Message()
        logc.Errorf(logCtx, "call auth error: (%d) %s", code, message)
    } else {
        code = GwErrCallAuth
        message = "call auth failed"
        logc.Errorf(logCtx, "%s: %s\n", message, err.Error())
    }

    // 出错响应
    logc.Errorf(logCtx, "Call auth failed: %s\n", err.Error())
    responseBytes, _ := NewResponseStr(logCtx, code, message, "")
    w.Header().Set("Content-Type", "application/json")
    w.Write(responseBytes)
}

func authHandleCallSuccess(next http.HandlerFunc, logCtx context.Context, w http.ResponseWriter, r *http.Request, authResp *mooonauth.AuthResp) {
    newReq := r.WithContext(r.Context())

    // 写 http 头
    for name, value := range authResp.HttpHeaders {
        // 写给调用者的
        w.Header().Set(name, value)

        // 写个被调服务的
        if strings.HasPrefix(name, "Grpc-Metadata-") {
            newReq.Header.Set(name, value)
        } else {
            newName := "Grpc-Metadata-" + name // 以 "Grpc-Metadata-" 打头的才能传递给被调服务，这是 go-zero 框架要求
            newReq.Header.Set(newName, value)
        }
    }

    logc.Infof(logCtx, "auth success: host=%s, path=%s, remote=%s", r.Host, r.URL.Path, r.RemoteAddr)
    next.ServeHTTP(w, newReq) // 鉴权成功，往下转发

    // 写 cookies
    for _, authCookie := range authResp.HttpCookies {
        httpCookie := AuthCookie2HttpCookie(authCookie)
        http.SetCookie(w, httpCookie)
    }
}
