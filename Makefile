# Makefile for Tharsis CLI

MODULE = $(shell go list -m)
VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || echo "1.0.0")
PACKAGES := $(shell go list ./... | grep -v /vendor/)
BINARY=tharsis
BUILD_PATH=$(MODULE)/cmd/tharsis
GCFLAGS:=-gcflags all=-trimpath=${PWD}
LDFLAGS := -ldflags "-X main.Version=${VERSION}"

## build the binaries
.PHONY: build
build:  ## build the Tharsis CLI binary
	CGO_ENABLED=0 go build ${LDFLAGS} -a -o ${BINARY} $(BUILD_PATH)

.PHONY: lint
lint: ## run golint on all Go package
	@echo "Checking go.mod..."
	@go mod tidy -diff > /dev/null
	@echo "Linting Go code..."
	@revive -set_exit_status $(PACKAGES)
	@echo "Checking Go formatting..."
	@UNFORMATTED=$$(gofmt -l . 2>/dev/null | grep -v vendor | grep -v testdata | grep -v '/pkg/mod/'); \
	if [ -n "$$UNFORMATTED" ]; then \
		echo "Files not formatted:"; \
		echo "$$UNFORMATTED"; \
		exit 1; \
	fi

.PHONY: vet
vet: ## run golint on all Go package
	@go vet $(PACKAGES)

.PHONY: fmt
fmt: ## run "go fmt" on all Go packages
	@go fmt $(PACKAGES)

.PHONY: generate
generate:
	go generate -v ./...

.PHONY: test
test: ## run unit tests
	go test -v ./...

release:
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_darwin_amd64  $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_darwin_arm64  $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=freebsd GOARCH=386   go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_freebsd_386   $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_freebsd_amd64 $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=freebsd GOARCH=arm   go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_freebsd_arm   $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=linux   GOARCH=386   go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_linux_386     $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_linux_amd64   $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm   go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_linux_arm     $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_linux_arm64   $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=openbsd GOARCH=386   go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_openbsd_386   $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=openbsd GOARCH=amd64 go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_openbsd_amd64 $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=solaris GOARCH=amd64 go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_solaris_amd64 $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=windows GOARCH=386   go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_windows_386   $(BUILD_PATH)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ${GCFLAGS} ${LDFLAGS} -a -o ./bin/${BINARY}_${VERSION}_windows_amd64 $(BUILD_PATH)

.PHONY: release-prep
release-prep: ## batch unreleased changie fragments into the changelog (VERSION=vX.Y.Z for final, VERSION=vX.Y.Z-alpha.1 for prerelease, or omit to auto-compute a final)
	@command -v changie >/dev/null 2>&1 || { echo "changie not found. Install: https://changie.dev/guide/installation/"; exit 1; }
	@set -eu; \
	if [ -z "$${VERSION:-}" ] && ls .changes/*-*.md >/dev/null 2>&1; then \
		echo "Prerelease version files exist in .changes/ — auto-compute would over-bump the version."; \
		echo "Pass VERSION=vX.Y.Z explicitly (the intended base/final version)."; \
		exit 1; \
	fi; \
	REL_VERSION=$${VERSION:-$$(changie next auto)}; \
	REL_VERSION=$${REL_VERSION#v}; \
	case "$$REL_VERSION" in \
		*-*) \
			BASE=$${REL_VERSION%%-*}; \
			PRERELEASE=$${REL_VERSION#*-}; \
			echo "Preparing prerelease changelog for v$$REL_VERSION"; \
			changie batch $$BASE --prerelease $$PRERELEASE --keep; \
			;; \
		*) \
			echo "Preparing changelog for v$$REL_VERSION"; \
			changie batch $$REL_VERSION --remove-prereleases; \
			;; \
	esac; \
	changie merge; \
	echo "✅ CHANGELOG.md updated."; \
	echo "   Commit the change. Once it lands on the default branch, CI tags the version and cuts the release."

# TEMPORARY — local test helper for validating the temp-branch prerelease model
# before the create-release CI job can be exercised (RELEASE_TOKEN is a protected
# variable unavailable in branch/MR pipelines). Remove once the CI prerelease
# path has been validated end-to-end.
# Mirrors the create-release CI job's prerelease path exactly:
# temp branch → commit changelog → GitLab API tag → delete temp branch.
# Requires: GITLAB_TOKEN (a personal access token with `api` scope)
#           VERSION       (e.g. VERSION=v0.36.0-alpha.2)
# Optional: ALLOW_NON_MAIN=1 to run from a non-main branch (needed pre-merge)
# Usage: GITLAB_TOKEN=<pat> VERSION=v0.36.0-alpha.2 ALLOW_NON_MAIN=1 make prerelease
.PHONY: prerelease
prerelease: ## TEMPORARY: test the temp-branch prerelease model locally (remove after CI path is validated)
	@command -v changie >/dev/null 2>&1 || { echo "changie not found. Install: https://changie.dev/guide/installation/"; exit 1; }
	@[ -n "$${GITLAB_TOKEN:-}" ] || { echo "GITLAB_TOKEN is required (a personal access token with api scope)"; exit 1; }
	@[ -n "$${VERSION:-}" ] || { echo "VERSION is required, e.g. VERSION=v0.36.0-alpha.2"; exit 1; }
	@case "$${VERSION}" in *-*) ;; *) echo "VERSION must be a prerelease (contain a hyphen)"; exit 1;; esac
	@if [ "$$(git rev-parse --abbrev-ref HEAD)" != "main" ] && [ -z "$${ALLOW_NON_MAIN:-}" ]; then \
		echo "Run from main, or set ALLOW_NON_MAIN=1 to override."; exit 1; fi
	@set -eu; \
	REL_VERSION=$${VERSION#v}; \
	TEMP_BRANCH="release-prep/local-$$REL_VERSION"; \
	ORIGIN_URL=$$(git remote get-url origin); \
	GITLAB_HOST="gitlab.com"; \
	PROJECT_PATH=$$(echo "$$ORIGIN_URL" | sed -E 's|.*:([^/].*\.git)$$|\1|;s|\.git$$||;s|/|%2F|g'); \
	echo "=== Creating temp branch $$TEMP_BRANCH ==="; \
	git checkout -b "$$TEMP_BRANCH"; \
	echo "=== Batching changelog ==="; \
	$(MAKE) release-prep VERSION=$$VERSION; \
	if [ -z "$$(git status --porcelain)" ]; then echo "No changes produced"; git checkout -; git branch -D "$$TEMP_BRANCH"; exit 1; fi; \
	git config user.name "release-bot"; \
	git config user.email "release-bot@noreply.$$GITLAB_HOST"; \
	git add CHANGELOG.md .changes/; \
	git commit -m "chore(release): $$VERSION"; \
	RELEASE_SHA=$$(git rev-parse HEAD); \
	echo "=== Pushing temp branch $$TEMP_BRANCH ==="; \
	git push "https://oauth2:$${GITLAB_TOKEN}@$$GITLAB_HOST/$$(echo "$$PROJECT_PATH" | sed 's|%2F|/|g').git" "$$TEMP_BRANCH"; \
	echo "=== Checking tag does not already exist ==="; \
	EXISTING=$$(curl -sf --header "PRIVATE-TOKEN: $${GITLAB_TOKEN}" \
		"https://$$GITLAB_HOST/api/v4/projects/$$PROJECT_PATH/repository/tags/$$VERSION" 2>/dev/null || echo ""); \
	if [ -n "$$EXISTING" ]; then \
		echo "Tag $$VERSION already exists — nothing to do."; \
		git push "https://oauth2:$${GITLAB_TOKEN}@$$GITLAB_HOST/$$(echo "$$PROJECT_PATH" | sed 's|%2F|/|g').git" --delete "$$TEMP_BRANCH" || true; \
		git checkout -; git branch -D "$$TEMP_BRANCH"; exit 0; fi; \
	echo "=== Creating tag $$VERSION at $$RELEASE_SHA via API ==="; \
	NOTES=$$(tail -n +2 ".changes/$$REL_VERSION.md" 2>/dev/null || echo ""); \
	curl --fail-with-body --request POST \
		--header "PRIVATE-TOKEN: $${GITLAB_TOKEN}" \
		"https://$$GITLAB_HOST/api/v4/projects/$$PROJECT_PATH/repository/tags" \
		--data-urlencode "tag_name=$$VERSION" \
		--data-urlencode "ref=$$RELEASE_SHA" \
		--data-urlencode "message=$$NOTES"; \
	echo ""; \
	echo "=== Deleting temp branch (prerelease: main untouched) ==="; \
	git push "https://oauth2:$${GITLAB_TOKEN}@$$GITLAB_HOST/$$(echo "$$PROJECT_PATH" | sed 's|%2F|/|g').git" --delete "$$TEMP_BRANCH" || true; \
	git checkout -; git branch -D "$$TEMP_BRANCH"; \
	echo "✅ Tag $$VERSION created. CI build/release pipeline should start now."; \
	echo "   Verify: the tagged commit's CHANGELOG.md shows ## $$VERSION at the top."

