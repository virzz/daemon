.PHONY: build
build:
	go build -o build/myservice ./daemon-cli/myservice/

clean:
	rm -rf build

test: build clean

install:
	go install ./daemon-cli/

test-local:
	go run ./daemon-cli/myservice/

test-remote:
	go run -tags remote ./daemon-cli/myservice/

