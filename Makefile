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

codegen: codegen-spec codegen-framework

codegen-spec:
	~/go/bin/tfplugingen-openapi generate \
		--config codegen/generator_config.yml \
		--output codegen/provider_code_spec.json \
		codegen/openapi_spec.yaml

codegen-framework:
	~/go/bin/tfplugingen-framework generate all \
		--input codegen/provider_code_spec.json \
		--output internal

update-openapi:
	# Pull the latest apic-generated spec from the ngrok-openapi repo
	cp ../ngrok-openapi/ngrok.yaml codegen/openapi_spec.yaml

dev: install
	@echo "Installed. Run 'rm -f test-manual/.terraform.lock.hcl && terraform -chdir=test-manual init' to test."

.PHONY: build install test testacc lint generate codegen codegen-spec codegen-framework update-openapi dev
