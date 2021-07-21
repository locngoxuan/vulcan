# exit when any command fails
set -e

CGO_ENABLED=0
GOOS=linux
GOARCH=amd64

go mod tidy

# build vulcan executor
go get -v ./cmd/vexec

rm -rf vendor

go mod vendor

go build --tags netgo -a -ldflags="-s -w" -o ./output/vulcan/toolchains/vexec ./cmd/vexec

# build vulcan set
go get -v ./cmd/vset

rm -rf vendor

go mod vendor

go build --tags netgo -a -ldflags="-s -w" -o ./output/vulcan/toolchains/vset ./cmd/vset

# build plugin: jfrog 
go get -v ./cmd/vset

rm -rf vendor

go mod vendor

go build --tags netgo -a -ldflags="-s -w" -o ./output/vulcan/plugins/jfrog ./plugins/jfrog
