build:
	go build -o build/myservice cmd/myservice/main.go

clean:
	rm -rf build

test: build clean