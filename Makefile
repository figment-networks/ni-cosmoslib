all: test


.PHONY: test
test:
	cd api && go test ./... -cover -race && cd ..


