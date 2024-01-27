# Written by yijian on 2024/01/21

all: mooon_gateway

mooon_gateway: main.go
	go build -o $@ $<

.PHONY: clean tidy

clean:
	rm -f mooon_gateway

tidy:
	go mod tidy
