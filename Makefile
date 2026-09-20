.PHONY: api worker event status

api:
	@go run ./tools/bridge api $(name) $(if $(table),--table=$(table))

worker:
	@go run ./tools/bridge worker $(name) $(if $(table),--table=$(table))

event:
	@go run ./tools/bridge event $(name) $(if $(table),--table=$(table))

status:
	@go run ./tools/bridge status
