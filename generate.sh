#!/bin/bash
protoc greet/proto/greet.proto --go_out=plugins=grpc:.
protoc blog/proto/blog.proto --go_out=plugins=grpc:.