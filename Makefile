GOARCH=amd64

build: linux darwin windows

linux: client-linux server-linux

darwin: client-darwin server-darwin

windows: client-windows server-windows

client-linux:
		cd client; GO111MODULE=on CGO_ENABLED=0 GOARCH=${GOARCH} GOOS=linux go build -o ../bin/client-linux-${GOARCH} .

client-darwin:
		cd client; GO111MODULE=on CGO_ENABLED=0 GOARCH=${GOARCH} GOOS=darwin go build -o ../bin/client-darwin-${GOARCH} .

client-windows:
		cd client; GO111MODULE=on CCGO_ENABLED=0 GOARCH=${GOARCH} GOOS=windows go build -o ../bin/client-windows-${GOARCH}.exe .

server-linux:
		cd server; GO111MODULE=on CCGO_ENABLED=0 GOARCH=${GOARCH} GOOS=linux go build -o ../bin/server-linux-${GOARCH} .

server-darwin:
		cd server; GO111MODULE=on CCGO_ENABLED=0 GOARCH=${GOARCH} GOOS=darwin go build -o ../bin/server-darwin-${GOARCH} .

server-windows:
		cd server; GO111MODULE=on CCGO_ENABLED=0 GOARCH=${GOARCH} GOOS=windows go build -o ../bin/server-windows-${GOARCH}.exe .

.PHONY: build
