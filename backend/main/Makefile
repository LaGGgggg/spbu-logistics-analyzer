.PHONY: test build vet lint fmt fmt-check mod-check check

build:
	go build ./...

vet:
	go vet ./... ../tests/...

lint:
	go tool staticcheck ./... ../tests/...

fmt:
	gofmt -w $$(find . ../tests -name '*.go' -type f -not -path './var/*')

fmt-check:
	@files="$$(gofmt -l $$(find . ../tests -name '*.go' -type f -not -path './var/*'))"; \
	if [ -n "$$files" ]; then printf '%s\n' "$$files"; exit 1; fi

mod-check:
	go mod tidy -diff
	go mod verify

test:
	go test -race ../tests/...

check: fmt-check mod-check build vet lint test
