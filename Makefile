# Written by yijian on 2024/01/21

all: mooon_gateway

mooon_gateway: main.go
	go build -o $@ $<

auth: mooon_auth.proto
	goctl rpc protoc mooon_auth.proto --go_out=./pb --go-grpc_out=./pb --zrpc_out=. --style go_zero;rm -fr internal mooon_auth.go

login: mooon_login.proto
	goctl rpc protoc mooon_login.proto --go_out=./pb --go-grpc_out=./pb --zrpc_out=. --style go_zero;rm -fr internal mooon_login.go

.PHONY: clean tidy

clean:
	rm -f mooon_gateway

tidy:
	go mod tidy
