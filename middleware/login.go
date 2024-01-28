package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc/status"
	"net/http"
	"net/url"
	"strings"

	"gateway/loginclient"
	"gateway/protoc/login"
)

// 登录服务客户端
var loginClient loginclient.Login

// NewLoginClient 实例化登录服务客户端
func NewLoginClient() {
	var loginConf zrpc.RpcClientConf

	conf.MustLoad("etc/login.yaml", &loginConf)
	// 注意如果这里失败会直接退出，因此在目标服务不可用时，总是退出。
	// 如果不想退出，使用 zrpc.NewClient 替代 zrpc.MustNewClient。
	// 实际上 zrpc.MustNewClient 只是包了下 zrpc.NewClient，增加了 Must 逻辑。
	client := zrpc.MustNewClient(loginConf)
	loginClient = loginclient.NewLogin(client)
}

// LoginMiddleware 登录
func LoginMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v1/") {
			next.ServeHTTP(w, r)
		} else {
			var me MyError
			var loginReq login.LoginReq

			params, _ := url.ParseQuery(r.URL.RawQuery)
			loginReq.Phone = params.Get("phone")
			loginReq.VerificationCode = params.Get("verification_code")
			fmt.Printf("Phone:%s, VerificationCode:%s\n", loginReq.Phone, loginReq.VerificationCode)
			loginResp, err := loginClient.Login(r.Context(), &loginReq)
			if err != nil {
				fmt.Println("login fail")
				//fmt.Fprintln(w, err)
				st, ok := status.FromError(err)
				if ok {
					fmt.Printf("%d", st.Code())
					me.Code = uint32(st.Code())
					me.Message = st.Message()
				} else {
					me.Code = 888899
					me.Message = err.Error()
				}
				jsonStr, _ := json.Marshal(&me)
				fmt.Fprintln(w, string(jsonStr))
			} else {
				cookie := &http.Cookie{ // 构造 cookie
					Name:  "mysid",
					Value: loginResp.SessionId,
					Path:  "/",
				}
				http.SetCookie(w, cookie) // 写 cookie
				//fmt.Fprintln(w, "Cookie has been set")
				me.Code = 0
				me.Message = "login success"
				jsonStr, _ := json.Marshal(&me)
				fmt.Fprintln(w, string(jsonStr))
			}
		}
	}
}
