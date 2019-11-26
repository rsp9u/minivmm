CC = gcc
BINARY = minivmm

GOOPTIONS=GOOS=$(GOOS) GOARCH=$(GOARCH) CC=${CC} CGO_ENABLED=1

.PHONY: all
all: statik
	${GOOPTIONS} go build -a -tags netgo -installsuffix netgo -o bin/${BINARY} cmd/main.go

statik: web/dist
	go get github.com/rakyll/statik
	go run github.com/rakyll/statik -src=./web/dist

web/dist:
	cd web; yarn install; yarn build

.PHONY: clean
clean:
	rm -rf bin web/dist statik
