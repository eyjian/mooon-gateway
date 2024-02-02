// Package middleware
// Written by yijian on 2024/01/28
package middleware

import (
	"context"
	"encoding/json"
	"github.com/zeromicro/go-zero/core/logc"
)

type Response struct {
    Code    uint32    `json:"code"`
    Message string `json:"message,omitempty"`
    Data    string `json:"data,omitempty"`
}

func NewResponseStr(ctx context.Context, code ErrCode, message string, data string) ([]byte, error) {
    var resp map[string]any

    // 尝试以 json 格式处理 data
    if len(data) > 0 {
        err := json.Unmarshal([]byte(data), &resp)
        if err != nil {
            logc.Errorf(ctx, "Unmarshal error: %s (%s)", err.Error(), data)
        } else {
            // 包装响应数据
            wrappedResp := map[string]any{
                "code":    code,
                "message": message,
                "data":    resp,
            }

            responseBytes, err := json.Marshal(wrappedResp)
            if err != nil {
                logc.Errorf(ctx, "Marshal error: %s (%s)", err.Error(), data)
            } else {
                return responseBytes, nil // data 为有效的 json 格式数据
            }
        }
    }

    // 以 json 格式处理 data 失败，转做普通字符串
    response := &Response{
        Code:    uint32(code),
        Message: message,
        Data:    data,
    }
    responseBytes, err := json.Marshal(response)
    if err != nil {
        logc.Errorf(ctx, "Marshal response error: %s", err.Error())
        return nil, err
    } else {
        return responseBytes, nil
    }
}