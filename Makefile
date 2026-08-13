build:
	go build ./cmd/tfplugingen-openapi

lint:
	golangci-lint run

fmt:
	gofmt -s -w -e .

test:
	go test $$(go list ./... | grep -v /output) -v -cover -timeout=120s -parallel=4

# Generate copywrite headers.
#
# Do not run this without checking what it would do first (append --plan to the
# go:generate directive in tools/copywrite.go). It generates no code -- it only
# stamps license headers -- and current copywrite releases default the copyright
# holder to "IBM Corp." following the HashiCorp acquisition, so running it here
# rewrites the header of every file in the repo. This fork keeps the upstream
# HashiCorp MPL-2.0 notices; new files should copy the two-line header from a
# neighbouring file by hand.
#
# The target also does not build as-is: this module's resolved knadh/koanf
# versions are incompatible with copywrite v0.25.3. Invoking copywrite with its
# own pins (go run github.com/hashicorp/copywrite@v0.25.3) works around that.
generate:
	cd tools; go generate ./...

# Regenerate testdata folder
testdata:
	go run ./cmd/tfplugingen-openapi generate \
		--config ./internal/cmd/testdata/petstore3/generator_config.yml \
		--output ./internal/cmd/testdata/petstore3/provider_code_spec.json \
		./internal/cmd/testdata/petstore3/openapi_spec.json

	go run ./cmd/tfplugingen-openapi generate \
		--config ./internal/cmd/testdata/github/generator_config.yml \
		--output ./internal/cmd/testdata/github/provider_code_spec.json \
		./internal/cmd/testdata/github/openapi_spec.json

	go run ./cmd/tfplugingen-openapi generate \
		--config ./internal/cmd/testdata/scaleway/generator_config.yml \
		--output ./internal/cmd/testdata/scaleway/provider_code_spec.json \
		./internal/cmd/testdata/scaleway/openapi_spec.yml

	go run ./cmd/tfplugingen-openapi generate \
		--config ./internal/cmd/testdata/edgecase/generator_config.yml \
		--output ./internal/cmd/testdata/edgecase/provider_code_spec.json \
		./internal/cmd/testdata/edgecase/openapi_spec.yml

	go run ./cmd/tfplugingen-openapi generate \
		--config ./internal/cmd/testdata/kubernetes/generator_config.yml \
		--output ./internal/cmd/testdata/kubernetes/provider_code_spec.json \
		./internal/cmd/testdata/kubernetes/openapi_spec.json

.PHONY: lint fmt test