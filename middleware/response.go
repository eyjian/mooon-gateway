// Package middleware
// Written by yijian on 2024/01/28
package middleware

import (
	"context"
	"encoding/json"
	"github.com/zeromicro/go-zero/core/logc"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func NewResponseStr(ctx context.Context, code int, message string, data any) []byte {
	response := &Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
	responseBytes, err := json.Marshal(response)
	if err != nil {
		logc.Errorf(ctx, "Marshal response error: %s", err.Error())
		return nil
	} else {
		return responseBytes
	}
}
