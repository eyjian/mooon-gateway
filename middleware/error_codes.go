// Package middleware
// Written by yijian on 2024/01/28
package middleware

const (
	GwSuccess      = 0
	GwErrSystem    = 202400002 // 系统错误（内部错误）
	GwErrUnknown   = 202400003 // 未知错误
	GwErrConnLogin = 202400004 // 连接登录服务出错
	GwErrConnAuth  = 202400005 // 连接鉴权服务出错
	GwErrCallLogin = 202400006 // 调用登录服务出错
	GwErrCallAuth  = 202400007 // 调用鉴权服务出错
)
