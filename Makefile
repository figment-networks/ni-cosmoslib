all: test


.PHONY: test
test:
	cd api && go test ./... -cover -race && cd ..
	cd util && go test ./... -cover -race && cd ..
	cd client && go test ./... -cover -race && cd ..
	cd flow && go test ./... -cover -race && cd ..
	cd ibcmapper && go test ./... -cover -race && cd ..
	cd ibcmapper/v2 && go test ./... -cover -race && cd ../..
	cd ibcmapper/v3 && go test ./... -cover -race && cd ../..


