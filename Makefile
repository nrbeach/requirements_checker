format:
	gofmt -w main.go

test:
	go test

setup:
	pre-commit install --install-hooks
