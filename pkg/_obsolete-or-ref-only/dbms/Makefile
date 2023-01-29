BINARY_NAME=pager

build:
	GOARCH=amd64 GOOS=darwin go build -o ${BINARY_NAME}-darwin main.go
	GOARCH=amd64 GOOS=linux go build -o ${BINARY_NAME}-linux main.go
	GOARCH=amd64 GOOS=window go build -o ${BINARY_NAME}-windows main.go

foo(arg):
	echo $arg

run:
	./${BINARY_NAME}

build_and_run: build run

clean:
	go clean
	rm ${BINARY_NAME}-darwin
	rm ${BINARY_NAME}-linux
	rm ${BINARY_NAME}-windows

test_files?=*_test.go

testz:
	echo "Running tests on $(testfiles)..."
	go test ./... $(testfiles)

test_race:
	go test ./... -race %(_test.go)

dep:
	go mod download

vet:
	go vet

files := file1 file2
some_file: $(files)
	echo "Look at this variable: " $(files)

tests := "*_test.go"
test:
	@echo "Running tests on ./... $(tests)"
	go test ./...



