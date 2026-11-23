IMAGE ?= ghcr.io/celestiaorg/ibc-go-simd
TAG ?= local

build-image:
	docker build --build-arg IBC_GO_VERSION=$(TAG) -t $(IMAGE):$(TAG) ..

init:
	./scripts/init.sh

e2e-test: init
	./scripts/run-e2e.sh "$(test)" "$(entrypoint)"

e2e-suite: init
	RUN_SUITE=true ./scripts/run-e2e.sh "" "$(entrypoint)"

# Compile and link every test binary without executing tests or starting Docker.
compile:
	go test -mod=readonly -tags=e2e -exec=true ./...

.PHONY: build-image init e2e-test e2e-suite compile
