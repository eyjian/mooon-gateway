// Package middleware
// Written by yijian on 2024/01/29
package middleware

import (
	"google.golang.org/grpc/status"
)

// NewGatewayError 创建网关代理的 rpc 服务返回的错误，网关不能正确处理其它的 error
func NewGatewayError(code codes.Code, message string) error {
	st := status.New(code, message)
	return st.Err()
}
