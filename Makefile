.PHONY: scaffold help

help:
	@echo "OSPA Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  scaffold    - Run the scaffold tool (use: make scaffold SERVICE=glance RESOURCES=image,member)"
	@echo "  build       - Build the agent"
	@echo "  test        - Run tests"
	@echo "  test-e2e    - Run e2e tests"

scaffold:
	@if [ -z "$(SERVICE)" ]; then \
		echo "Error: SERVICE is required"; \
		echo "Usage: make scaffold SERVICE=glance RESOURCES=image,member [DISPLAY_NAME=Glance] [TYPE=image]"; \
		exit 1; \
	fi
	@go run ./cmd/scaffold \
		--service $(SERVICE) \
		--display-name $(or $(DISPLAY_NAME),$(shell echo $(SERVICE) | sed 's/^./\U&/')) \
		--resources $(RESOURCES) \
		--type $(or $(TYPE),$(SERVICE)) \
		$(if $(FORCE),--force,)

build:
	go build -o bin/ospa-agent ./cmd/agent

test:
	go test ./...

test-e2e:
	go test -tags=e2e ./e2e/...

