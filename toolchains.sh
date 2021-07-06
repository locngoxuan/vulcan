# exit when any command fails
set -e

go mod tidy

# build vulcan executor
go get -v ./cmd/vexec

rm -rf vendor

go mod vendor

go build -ldflags="-s -w" -o ./output/vulcan/toolchains/vexec ./cmd/vexec

# build vulcan set
go get -v ./cmd/vset

rm -rf vendor

go mod vendor

go build -ldflags="-s -w" -o ./output/vulcan/toolchains/vset ./cmd/vset
