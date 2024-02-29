.PHONY: build
build:
	go build -o build/myservice ./daemon-cli/myservice/main.go

clean:
	rm -rf build

test: build clean

install:
	go install ./daemon-cli/