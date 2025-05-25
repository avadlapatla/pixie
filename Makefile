.PHONY: dev down lint install-golangci-lint proto plugins ui-deps ui-build generate-keys switch-hs256 switch-rs256

dev: ui-build
	docker compose -f deployments/docker-compose.yml up --build

down:
	docker compose -f deployments/docker-compose.yml down

generate-keys:
	@mkdir -p keys
	@openssl genrsa -out keys/private.pem 2048
	@openssl rsa -in keys/private.pem -pubout -out keys/public.pem
	@chmod 600 keys/private.pem
	@chmod 644 keys/public.pem
	@echo "RSA keys generated in keys/ directory"

switch-hs256:
	@grep -v "JWT_ALGO\|JWT_SECRET\|JWT_PUBLIC_KEY_FILE\|JWT_PRIVATE_KEY_FILE" .env > .env.tmp || touch .env.tmp
	@echo "JWT_ALGO=HS256" >> .env.tmp
	@echo "JWT_SECRET=supersecret123" >> .env.tmp
	@mv .env.tmp .env
	@echo "Switched to HS256 algorithm"

switch-rs256: generate-keys
	@grep -v "JWT_ALGO\|JWT_SECRET\|JWT_PUBLIC_KEY_FILE\|JWT_PRIVATE_KEY_FILE" .env > .env.tmp || touch .env.tmp
	@echo "JWT_ALGO=RS256" >> .env.tmp
	@echo "JWT_PUBLIC_KEY_FILE=$(shell pwd)/keys/public.pem" >> .env.tmp
	@echo "JWT_PRIVATE_KEY_FILE=$(shell pwd)/keys/private.pem" >> .env.tmp
	@mv .env.tmp .env
	@echo "Switched to RS256 algorithm"

lint: install-golangci-lint
	cd core && golangci-lint run ./...

install-golangci-lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && \
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.55.2)

proto:
	buf generate

plugins-noop:
	cd plugins/noop && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../plugin-noop .

plugins-thumb:
	cd plugins/thumbnailer && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ../plugin-thumbnailer .

plugins: plugins-thumb plugins-noop

ui-deps:
	cd plugins/ui-react && npm install

ui-build: ui-deps
	cd plugins/ui-react && npm run build
