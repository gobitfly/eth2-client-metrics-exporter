GITCOMMIT=`git describe --always`
GITDATE=`TZ=UTC git show -s --date=iso-strict-local --format=%cd HEAD`
GITDATESHORT=$$(TZ=UTC git show -s --date=iso-strict-local --format=%cd HEAD | sed 's/[-T:]//g' | sed 's/\(+.*\)$$//g')
BUILDDATE=`date -u +"%Y-%m-%dT%H:%M:%S%:z"`
BUILDDATESHORT=`date -u +"%Y%m%d%H%M%S"`
VERSION=${GITDATESHORT}-${GITCOMMIT}
PACKAGE=github.com/guybrush/graffitiwallpainter
LDFLAGS="-X main.Version=${VERSION} -X main.BuildDate=${BUILDDATE} -X main.GitCommit=${GITCOMMIT} -X main.GitDate=${GITDATE} -X main.GitDateShort=${GITDATESHORT}"
DOCKERIMAGE="gobitfly/eth2-client-metrics-exporter"
BINARY=bin/eth2-client-metrics-exporter
all: test build
test:
	go test -v ./...
clean:
	rm -rf bin
build:
	go build --ldflags=${LDFLAGS} -o ${BINARY}
build-docker:
	docker build -t ${DOCKERIMAGE} -t ${DOCKERIMAGE}:${GITDATESHORT}-${GITCOMMIT} .
push-docker:
	docker push ${DOCKERIMAGE}
	docker push ${DOCKERIMAGE}:${GITDATESHORT}-${GITCOMMIT}
