# Written by yijian on 2024/01/21

all: mooon_gateway

mooon_gateway: main.go middleware/auth.go middleware/login.go
	go build -o $@ $<

auth: mooon_auth.proto
	goctl rpc protoc mooon_auth.proto --go_out=./pb --go-grpc_out=./pb --zrpc_out=. --style go_zero;rm -fr internal mooon_auth.go etc/mooon_auth.yaml

login: mooon_login.proto
	goctl rpc protoc mooon_login.proto --go_out=./pb --go-grpc_out=./pb --zrpc_out=. --style go_zero;rm -fr internal mooon_login.go etc/mooon_login.yaml

.PHONY: clean tidy fetch

clean:
	rm -f mooon_gateway

tidy:
	go mod tidy

fetch: # 强制用远程仓库的覆盖本地，运行时需指定分支名，如：make fetch branch=main
	git fetch --all&&git reset --hard origin/$$branch
