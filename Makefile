.PHONY: test test-integration verify images

test:
	go test ./...

# Richiede Postgres e Redis raggiungibili su localhost; sovrascrivi gli indirizzi con le due variabili.
test-integration:
	LINKBOARD_TEST_DATABASE_URL=$${LINKBOARD_TEST_DATABASE_URL:-postgres://linkboard:linkboard@localhost:5432/linkboard?sslmode=disable} \
	LINKBOARD_TEST_REDIS_ADDR=$${LINKBOARD_TEST_REDIS_ADDR:-localhost:6379} \
	go test -count=1 ./...

verify:
	go vet ./...
	@test -z "$$(gofmt -l .)" || (echo "file non formattati:"; gofmt -l .; exit 1)
	go test ./...

VERSION ?= v0.1.0
images:
	docker build --build-arg SERVICE=api -t linkboard-api:$(VERSION) .
	docker build --build-arg SERVICE=worker -t linkboard-worker:$(VERSION) .
	docker build -t linkboard-frontend:$(VERSION) frontend
