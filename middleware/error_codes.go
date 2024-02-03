// Package middleware
// Written by yijian on 2024/01/28
package middleware

type ErrCode uint32

const (
    GwSuccess      ErrCode = 0
    GwErrSystem    ErrCode = 202400002 // 系统错误（内部错误）
    GwErrUnknown   ErrCode = 202400003 // 未知错误
    GwErrConnLogin ErrCode = 202400004 // 连接登录服务出错
    GwErrConnAuth  ErrCode = 202400005 // 连接鉴权服务出错
    GwErrCallLogin ErrCode = 202400006 // 调用登录服务出错
    GwErrCallAuth  ErrCode = 202400007 // 调用鉴权服务出错
    GwInvalidResp  ErrCode = 202400008
)
