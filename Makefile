
.PHONY: run clean

main: main.go
	go build main.go

run:
	./main

clean:
	rm main

build:
	docker image build --tag=gotest .
