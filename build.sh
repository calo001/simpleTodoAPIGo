#!/usr/bin/env bash
go get -u github.com/gin-gonic/gin
go get -u github.com/appleboy/gin-jwt
go get -u github.com/lib/pq
go get -u github.com/lib/pq/hstore
go build -o ./app main.go