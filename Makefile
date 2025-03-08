.PHONY: lint test vendor clean

export GO111MODULE=on

# Убедимся, что golangci-lint установлен
GOLANGCI_LINT := $(shell which golangci-lint)

# Если golangci-lint не найден, то устанавливаем его
lint:
	@if [ -z "$(GOLANGCI_LINT)" ]; then \
		echo "golangci-lint не найден. Устанавливаю..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.33.0; \
	fi
	$(GOLANGCI_LINT) run

test:
	go test -v -cover ./...

yaegi_test:
	yaegi test -v .

vendor:
	go mod vendor

clean:
	rm -rf ./vendor
