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
                responseBytes, err := NewResponseStr(logCtx, GwErrConnAuth, "conn auth error", "")
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

            authReq.Body = string(reqBodyBytes)
            authResp, err := mooonAuth.Authenticate(r.Context(), &authReq)
            if err != nil {
                // 鉴权失败或者未通过
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

                return
            } else {
                // 鉴权通过，改写请求以加入（传递）鉴权数据
                newReq := r.WithContext(r.Context())

                // 写 http 头
                for name, value := range authResp.HttpHeaders {
                    newReq.Header.Set(name, value)
                }
                // 写 cookies
                for _, authCookie := range authResp.HttpCookies {
                    httpCookie := AuthCookie2HttpCookie(authCookie)
                    newReq.AddCookie(httpCookie)
                }

                // 往下转发
                next.ServeHTTP(w, newReq)
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
