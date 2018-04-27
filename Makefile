BINARYNAME=go-blog
TMP=tmp
GITHASH=$(shell git rev-parse HEAD)
BUILD_VERSION=$(git describe --tags)
BUILD_DATE=$(shell date -u '+%Y-%m-%d %I:%M:%S %Z')

RELEASE="releases"

LDFLAGS=-ldflags '-X main.BuildVersion=${BUILD_VERSION} -X main.GitHash=${GITHASH}'

.PHONY: clean build-release build lint install package vet fmt test

build-release: build package

build:
	go build ${LDFLAGS} -o ${GOPATH}/bin/go-blog
	cd clt/createuser && go build -o ${GOPATH}/bin/create_user ${LDFLAGS}
	cd clt/initdatabase && go build -o ${GOPATH}/bin/init_database ${LDFLAGS}

install:
	go install ${LDFLAGS}
	cd clt/createuser && go install ${LDFLAGS}
	cd clt/initdatabase && go install ${LDFLAGS}

package:
	rm -r ${TMP}
	mkdir -p ${TMP}/clt
	cp ${GOPATH}/bin/go-blog ${TMP}/
	cp ${GOPATH}/bin/create_user  ${TMP}/clt
	cp ${GOPATH}/bin/init_database ${TMP}/clt
	cp go-blog.conf ${TMP}/
	cp -r scripts/ ${TMP}/
	cp -r templates/ ${TMP}/
	cp -r assets/ ${TMP}/

vet:
	go vet ./...

fmt:
	go fmt ./...

test:
	go test -v ./...

clean:
	go clean -i ./...

lint:
	golint ./...
