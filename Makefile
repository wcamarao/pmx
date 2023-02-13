.PHONY: coverage
coverage:
	@mkdir -p .coverage
	@go test -count=1 -coverpkg=./... -coverprofile .coverage/report.out . > /dev/null
	@go tool cover -func=.coverage/report.out -o=.coverage/report.text
	@go tool cover -html=.coverage/report.out -o=.coverage/report.html
	@cat .coverage/report.text

.PHONY: compose
compose:
	@docker-compose up --detach

.PHONY: lint
lint:
	@go mod verify
	@golangci-lint run

.PHONY: test
test:
	@go test -count=1 ./...
