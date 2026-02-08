GOLANG_CI_LINT_VERSION="v2.0.2"

BIN_NAME=terraform-provider-taskmate
CONTAINER_NAME=${BIN_NAME}

CURRENT_DIR=$(shell pwd)
DIST_DIR=${CURRENT_DIR}/bin

.PHONY: all
all: build

.PHONY: build
build:
	go get .
	CGO_ENABLED=0 go build -o ${DIST_DIR}/${BIN_NAME} ./

.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/hashitalk.com/edu/taskmate/1.0.0/$$(go env GOOS)_$$(go env GOARCH)
	cp ${DIST_DIR}/${BIN_NAME} ~/.terraform.d/plugins/hashitalk.com/edu/taskmate/1.0.0/$$(go env GOOS)_$$(go env GOARCH)/

.PHONY: test
test:
	go test ./... -v -cover -timeout=120s -parallel=4

.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v -cover -timeout 120m

.PHONY: lint
lint:
	golangci-lint run -v

.PHONY: lint.docker
lint.docker:
	docker run -t --rm \
		-v $(shell pwd):/app \
		-w /app \
		golangci/golangci-lint:${GOLANG_CI_LINT_VERSION} \
		golangci-lint run -v

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: docs
docs:
	go get .
	go generate ./...

.PHONY: clean
clean:
	go clean
	rm -rf ${DIST_DIR}

.PHONY: run
run: build
	${DIST_DIR}/${BIN_NAME}
