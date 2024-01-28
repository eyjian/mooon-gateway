package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc/status"
	"net/http"
	"strings"

	"gateway/authclient"
	"gateway/protoc/auth"
)

// AuthMiddleware 鉴权
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v2/") {
			next.ServeHTTP(w, r)
		} else {
			var me MyError

			cookie, err := r.Cookie("mysid")
			if err != nil {
				me.Code = 666699
				me.Message = "no access in gateway.AuthMiddleware"
				jsonStr, _ := json.Marshal(&me)
				fmt.Fprintln(w, string(jsonStr))
			} else {
				// cookies 中有会话 ID
				var authReq auth.AuthReq
				var authConf zrpc.RpcClientConf

				conf.MustLoad("etc/auth.yaml", &authConf)
				client := zrpc.MustNewClient(authConf)
				authClient := authclient.NewAuth(client)

				authReq.SessionId = cookie.Value
				authResp, err := authClient.Authenticate(r.Context(), &authReq) // 调用鉴权服务
				if err != nil {
					// 未通过鉴权
					st, ok := status.FromError(err)
					if ok {
						me.Code = uint32(st.Code())
						me.Message = st.Message()
					} else {
						me.Code = 999988
						me.Message = err.Error()
					}
					jsonStr, _ := json.Marshal(&me)
					fmt.Fprintln(w, string(jsonStr))
				} else {
					// 通过鉴权
					newReq := r.WithContext(r.Context())
					newReq.Header.Set("Grpc-Metadata-myuid", authResp.UserId)

					// 往下转发
					next.ServeHTTP(w, newReq)
				}
			}
		}
	}
}
