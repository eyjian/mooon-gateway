# mooon-gateway

基于 go-zero 实现的 API 网关

# 出错代码

网关的出错代码为 **9** 位数的数字值，并且总是以 **2024** 打头，基于网关的服务应对避免出现相同值。网关用到的出错代码参见 [https://github.com/eyjian/mooon-gateway/blob/main/middleware/error_codes.go](https://github.com/eyjian/mooon-gateway/blob/main/middleware/error_codes.go) 。

# 如何开发鉴权服务

通常鉴权服务中鉴权通过后，会中请求中加入额外数据传递给被调服务。对于 http 转 gRPC，这部分数据通过 http 头传递，被调服务可通过 metadata.ValueFromIncomingContext 取得。
