.PHONY: api worker event status test cover migrate migrate-down migrate-status migration table

api:
	@go run ./tools/bridge api $(name) $(if $(table),--table=$(table))

worker:
	@go run ./tools/bridge worker $(name) $(if $(table),--table=$(table))

event:
	@go run ./tools/bridge event $(name) $(if $(table),--table=$(table))

status:
	@go run ./tools/bridge status

migrate:
	@go run ./cmd/migrate up

migrate-down:
	@go run ./cmd/migrate down

migrate-status:
	@go run ./cmd/migrate status

migration:
	@go run ./tools/bridge migration $(name)

table:
	@go run ./tools/bridge table $(name)

test:
	@go test ./...

cover:
	@go test ./... -coverprofile=coverage.out
	@go tool cover -func=coverage.out | tail -1
