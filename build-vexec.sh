# exit when any command fails
set -e

go mod tidy

go get -v ./cmd/vexec

rm -rf vendor

go mod vendor

go build -ldflags="-s -w" -o ./output/vulcan/toolchains/vexec ./cmd/vexec
