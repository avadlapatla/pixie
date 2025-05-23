.PHONY: dev down lint install-golangci-lint proto plugins

dev: plugins
	docker compose -f deployments/docker-compose.yml up --build

down:
	docker compose -f deployments/docker-compose.yml down

lint: install-golangci-lint
	cd core && golangci-lint run ./...

install-golangci-lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && \
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.55.2)

proto:
	buf generate

plugins-auth:
	cd plugins/authjwt && go build -o ../plugin-authjwt .

plugins-noop:
	cd plugins/noop && go build -o ../plugin-noop .

plugins: plugins-auth plugins-noop   # ensure existing noop still builds
