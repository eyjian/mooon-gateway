package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/status"
	"mooon-gateway/middleware"
	"net/http"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/gateway"
)

var configFile = flag.String("f", "etc/gateway.yaml", "the config file")

func main() {
	flag.Parse()

	conf.MustLoad(*configFile, &middleware.GlobalConfig)
	server := gateway.MustNewServer(middleware.GlobalConfig.GatewayConf)
	server.Use(middleware.LoginMiddleware)
	server.Use(middleware.AuthMiddleware)
	server.Use(wrapResponse)
	defer server.Stop()

	// 设置错误处理
	httpx.SetErrorHandler(grpcErrorHandler)

	fmt.Printf("Starting mooon_gateway at %s:%d...\n", middleware.GlobalConfig.Host, middleware.GlobalConfig.Port)
	server.Start()
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(p []byte) (int, error) {
	return rw.body.Write(p)
}

func (rw *responseWriter) Body() []byte {
	return rw.body.Bytes()
}

// 对响应加上“"code":0,"data":{}”，
// 对于已经包含了“code”的不做任何处理（原因是 grpcErrorHandler 才能处理好）
func wrapResponse(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logCtx := logx.ContextWithFields(r.Context(), logx.Field("path", r.URL.Path))

		// 记录原始响应 writer
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// 执行下一个中间件或处理函数
		next.ServeHTTP(rw, r)

		// 检查响应状态码
		if rw.statusCode != http.StatusOK {
			return
		}

		// 获取原始响应数据
		var resp map[string]interface{}
		err := json.Unmarshal(rw.Body(), &resp)
		if err != nil {
			logc.Errorf(logCtx, "Unmarshal response error: %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 检查响应是否已经包含 code
		if _, ok := resp["code"]; ok {
			// 如果响应已经包含 code，则直接写回原始响应正文
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write(rw.Body())
			if err != nil {
				logc.Errorf(logCtx, "Write response error: %s\n", err.Error())
			}
			return
		}

		// 包装响应数据
		wrappedResp := map[string]interface{}{
			"code":    0,
			"message": "success",
			"data":    resp,
		}

		// 将包装后的响应数据写回 response body
		//w.Header().Set("Content-Type", "application/json")
		//json.NewEncoder(w).Encode(wrappedResp)
		httpx.OkJson(w, wrappedResp) // 这里的实现不要有调用 httpx.SetOkHandler
	})
}

func grpcErrorHandler(err error) (int, any) {
	if st, ok := status.FromError(err); ok {
		return http.StatusOK, middleware.Response{
			Code:    int(st.Code()),
			Message: st.Message(),
		}
	}

	code := middleware.GwErrUnknown
	return http.StatusOK, middleware.Response{
		Code:    code,
		Message: err.Error(),
	}
}
