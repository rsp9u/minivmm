CC = gcc
BINARY = minivmm

GOOPTIONS=GOOS=$(GOOS) GOARCH=$(GOARCH) CC=${CC} CGO_ENABLED=1

.PHONY: all
all: web/dist
	${GOOPTIONS} go build -a -tags netgo -installsuffix netgo -o bin/${BINARY} cmd/main.go

web/dist:
	cd web; yarn install; yarn build

.PHONY: clean
clean:
	rm -rf bin web/dist
