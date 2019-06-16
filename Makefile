clean:
	rm -f fallback.cov

cover:
	go test ./... -coverprofile=fallback.cov
	go tool cover -html=fallback.cov

test:
	go test ./...