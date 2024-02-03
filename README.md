# mooon-gateway
基于 go-zero 实现的 API 网关


# 如何开发鉴权服务

通常鉴权服务中鉴权通过后，会中请求中加入额外数据传递给被调服务。对于 http 转 gRPC，这部分数据通过 http 头传递，被调服务可通过 metadata.ValueFromIncomingContext 取得。
