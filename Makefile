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

.PHONY: go-mod-tidy
go-mod-tidy:
	cd api && go mod tidy && cd ..
	cd util && go mod tidy && cd ..
	cd client && go mod tidy && cd ..
	cd flow && go mod tidy && cd ..
	cd ibcmapper && go mod tidy && cd ..
	cd ibcmapper/v2 && go mod tidy && cd ../..
	cd ibcmapper/v3 && go mod tidy && cd ../..
