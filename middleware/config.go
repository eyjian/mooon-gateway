// Package middleware
// Written by yijian on 2024/02/02
package middleware

import (
    "github.com/zeromicro/go-zero/core/discov"
    "github.com/zeromicro/go-zero/gateway"
    "github.com/zeromicro/go-zero/zrpc"
)

var GlobalConfig Config

// Config 网关配置文件
type Config struct {
    Etcd                discov.EtcdConf
    gateway.GatewayConf // 网关配置文件

    Auth struct {
        zrpc.RpcClientConf        // 鉴权服务客户端配置
        Prefix             string // 鉴权的 path 前缀，如：www.mooon.com/v1 中的 v1
    }

    Login struct {
        zrpc.RpcClientConf        // 鉴权服务服务端配置
        Prefix             string // 登录的 path 前缀
    }
}
