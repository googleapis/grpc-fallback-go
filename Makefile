clean:
	rm -f fallback.cov
	rm -f *.tar.gz

cover:
	go test ./... -coverprofile=fallback.cov
	go tool cover -html=fallback.cov

image:
	docker build -t gcr.io/gapic-images/fallback-proxy .

install:
	go install ./client
	go install ./server
	go install ./cmd/fallback-proxy

release:
	./release.sh

test:
	go test ./...