#!/bin/bash
#
# Copyright 2019 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

VERSION=0.1.0
GO111MODULE=on
GOPROXY=https://proxy.golang.org

go mod download

# linux-amd64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ./cmd/fallback-proxy
tar -czf grpc-fallback-go-$VERSION-linux-amd64.tar.gz fallback-proxy
rm -f fallback-proxy

# linux-arm
CGO_ENABLED=0 GOOS=linux GOARCH=arm go build ./cmd/fallback-proxy
tar -czf grpc-fallback-go-$VERSION-linux-arm.tar.gz fallback-proxy
rm -f fallback-proxy

# darwin-amd64
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build ./cmd/fallback-proxy
tar -czf grpc-fallback-go-$VERSION-darwin-amd64.tar.gz fallback-proxy
rm -f fallback-proxy

# windows-amd64
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ./cmd/fallback-proxy
tar -czf grpc-fallback-go-$VERSION-windows-amd64.tar.gz fallback-proxy.exe
rm -f fallback-proxy.exe

# build & tag image
make image
docker tag \
  gcr.io/gapic-images/fallback-proxy \
  gcr.io/gapic-images/fallback-proxy:$VERSION