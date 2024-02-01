// Package middleware
// Written by yijian on 2024/02/02
package middleware

import (
	"github.com/zeromicro/go-zero/gateway"
	"github.com/zeromicro/go-zero/zrpc"
)

var GlobalConfig Config

// Config 网关配置文件
type Config struct {
	gateway.GatewayConf                    // 网关配置文件
	AuthConf            zrpc.RpcClientConf // 鉴权服务客户端配置
	LoginConf           zrpc.RpcClientConf // 鉴权服务服务端配置
}
