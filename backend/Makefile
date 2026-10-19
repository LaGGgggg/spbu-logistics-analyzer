.PHONY: test build vet lint fmt fmt-check mod-check check

build:
	go build ./...

vet:
	go vet ./...

lint:
	go tool staticcheck ./...

fmt:
	gofmt -w $$(find cmd internal tests -name '*.go' -type f)

fmt-check:
	@files="$$(gofmt -l $$(find cmd internal tests -name '*.go' -type f))"; \
	if [ -n "$$files" ]; then printf '%s\n' "$$files"; exit 1; fi

mod-check:
	go mod tidy -diff
	go mod verify

test:
	go test -race ./tests/...

check: fmt-check mod-check build vet lint test
