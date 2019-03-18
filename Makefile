BINARYNAME=go-blog
TMP=tmp
GITHASH=$(shell git rev-parse HEAD)
BUILD_VERSION=$(shell git describe --tags)

RELEASE="releases"

LDFLAGS=-ldflags '-X main.BuildVersion=${BUILD_VERSION} -X main.GitHash=${GITHASH}'

.PHONY: clean build-release build lint install package vet fmt test

build-release: clean tidy fmt vet test build package

build:
	go build ${LDFLAGS} -o ${GOPATH}/bin/go-blog
	cd clt/createuser && go build -o ${GOPATH}/bin/create_user ${LDFLAGS}
	cd clt/initdatabase && go build -o ${GOPATH}/bin/init_database ${LDFLAGS}

install:
	go install ${LDFLAGS}
	cd clt/createuser && go install ${LDFLAGS}
	cd clt/initdatabase && go install ${LDFLAGS}

package:
	-rm -r ${TMP}
	mkdir -p ${TMP}/clt
	-mkdir -p releases
	cp ${GOPATH}/bin/go-blog ${TMP}/
	cp ${GOPATH}/bin/create_user ${TMP}/clt
	cp ${GOPATH}/bin/init_database ${TMP}/clt
	cp go-blog.conf ${TMP}/
	cp -r examples/ ${TMP}/
	cp -r templates/ ${TMP}/
	cp -r assets/ ${TMP}/
	cd ${TMP} && tar -czvf ../releases/$(BINARYNAME)_$(BUILD_VERSION).tar.gz * && cd -

vet:
	go vet ./...

fmt:
	go fmt ./...

test:
	go test ./...

clean:
	go clean -i ./...

tidy:
	go mod tidy
