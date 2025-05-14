.PHONY: example
example: clean
	env
	mkdir -p ./build && cd ./build && \
	go build ../example/default && \
	go build -tags remote ../example/remote && \
	ls -al ./

clean:
	rm -f ./build/*
