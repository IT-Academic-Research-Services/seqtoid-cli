export GO111MODULE=on
export CGO_ENABLED=1

test: ## run tests, will update go.mod
	go test -coverprofile=coverage.txt -covermode=atomic ./...
.PHONY: test

build: ## build the binary
	go build -o seqtoid -ldflags="-X github.com/IT-Academic-Research-Services/seqtoid-cli/pkg/auth0.defaultClientID=${AUTH0_CLIENT_ID} -X github.com/IT-Academic-Research-Services/seqtoid-cli/pkg/auth0.defaultAuth0Host=${AUTH0_HOST} -X github.com/IT-Academic-Research-Services/seqtoid-cli/pkg/auth0.defaultAudience=${AUTH0_AUDIENCE} -X github.com/IT-Academic-Research-Services/seqtoid-cli/pkg/seqtoid.defaultBaseURL=${SEQTOID_BASE_URL} -X github.com/IT-Academic-Research-Services/seqtoid-cli/pkg.Version=${VERSION}" .
.PHONY: build

deps:
	go mod tidy
.PHONY: deps

check-mod:
	go mod tidy
	git diff --exit-code -- go.mod go.sum
.PHONY: check-mod
