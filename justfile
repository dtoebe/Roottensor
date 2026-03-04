test *args="-v -cover":
  go test {{ args }} ./...

build:
  go build -v -o ./dist/RootTensor

tidy:
  go mod tidy

air:
  air

run: build
  ./dist/RootTensor

vet: tidy
  go vet ./...

clean:
  rm -rf ./dist
  rm -rf ./tmp

default:
  just --list
