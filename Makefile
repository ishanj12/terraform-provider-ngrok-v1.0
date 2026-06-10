default: build

build:
	go build -o terraform-provider-ngrok .

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/ngrok/ngrok/0.0.1/$(shell go env GOOS)_$(shell go env GOARCH)
	cp terraform-provider-ngrok ~/.terraform.d/plugins/registry.terraform.io/ngrok/ngrok/0.0.1/$(shell go env GOOS)_$(shell go env GOARCH)/

test:
	go test ./... -v

testacc:
	TF_ACC=1 go test ./... -v -timeout 120m

lint:
	golangci-lint run ./...

generate:
	go generate ./...

dev: install
	@echo "Installed. Run 'rm -f test-manual/.terraform.lock.hcl && terraform -chdir=test-manual init' to test."

.PHONY: build install test testacc lint generate dev
