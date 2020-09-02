
.PHONY: run clean

main: main.go
	go build

run:
	./main

clean:
	rm main

build:
	docker image build --tag=gotest .
