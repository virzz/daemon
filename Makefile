.PHONY: example
example: clean
	go build -o default.out ./example/default
	go build -tags remote -o remote.out ./example/remote
	ls -al *.out

clean: *.out
	rm -f *.out