.PHONY: api wire test test-integration

api:
	@./scripts/gen_openapi.sh

wire:
	@cd control/cmd/control && wire gen .

test:
	@cd common && go test ./...
	@cd control && go test ./...
	@cd offline && go test ./...
	@cd online && go test ./...

test-int:
	@cd common && go test -tags=integration ./registry
