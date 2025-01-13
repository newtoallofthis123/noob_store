BINARY_NAME=noob_store

all: run

build: 
	@cd cmd && go build -o ../bin/$(BINARY_NAME)

run: build
	@./bin/$(BINARY_NAME)

clean:
	@rm -f bin/$(BINARY_NAME)
