GITCOMMIT=`git describe --always`
GITDATE=`TZ=UTC git show -s --date=iso-strict-local --format=%cd HEAD`
LDFLAGS="-X main.GitCommit=${GITCOMMIT} -X main.GitDate=${GITDATE}"
BINARY=bin/eth2-client-metrics-exporter
all: test build
test:
	go test -v ./...
clean:
	rm -rf bin
build:
	./build.sh
build-binary-linux-amd64:
	env GOOS=linux GOARCH=amd64 go build --ldflags=${LDFLAGS} -o ${BINARY}-linux-amd64
build-binary-linux-arm64:
	env GOOS=linux GOARCH=arm64 go build --ldflags=${LDFLAGS} -o ${BINARY}-linux-arm64
