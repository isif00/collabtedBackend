.PHONY: prisma-generate prisma-db-push tidy build

prisma-generate:
	go run github.com/steebchen/prisma-client-go generate --schema ./prisma

prisma-db-push:
	go run github.com/steebchen/prisma-client-go db push --schema ./prisma

tidy:
	go mod tidy

build:
	go build -o bin/ap ./cmd/server/main.go
