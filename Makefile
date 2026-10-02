.PHONY: verify test test-integration test-browser
verify:
	python3 scripts/check-source.py
	python3 -m unittest discover -s scripts -p 'test_*.py'
	cd backend && test -z "$$(gofmt -l cmd internal)" && go vet ./... && go test -race ./... && go build ./cmd/dedhed
	pnpm --dir web install --frozen-lockfile
	pnpm --dir web typecheck
	pnpm --dir web lint
	pnpm --dir web test
	pnpm --dir web build
test:
	cd backend && go test -race ./...
	pnpm --dir web test
test-integration:
	test -n "$(TEST_DATABASE_URL)"
	cd backend && go test -race ./internal/catalog -count=1
test-browser:
	pnpm --dir web test:browser
