.PHONY: api worker event status test cover migrate migrate-down migrate-status migration table

api:
	@go run ./tools/junction api $(name) $(if $(table),--table=$(table))

worker:
	@go run ./tools/junction worker $(name) $(if $(table),--table=$(table))

event:
	@go run ./tools/junction event $(name) $(if $(table),--table=$(table))

status:
	@go run ./tools/junction status

migrate:
	@go run ./cmd/migrate up

migrate-down:
	@go run ./cmd/migrate down

migrate-status:
	@go run ./cmd/migrate status

migration:
	@go run ./tools/junction migration $(name)

table:
	@go run ./tools/junction table $(name)

test:
	@go test ./...

cover:
	@go test ./... -coverprofile=coverage.out
	@go tool cover -func=coverage.out | tail -1
