#!/usr/bin/env bash

protoc -I/usr/local/include -I. \
  -I${GOPATH}/src \
  -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  -I${GOPATH}/src/github.com/envoyproxy/protoc-gen-validate \
  --grpc-gateway_out=logtostderr=true:. \
  *.proto