#!/bin/bash

export GOPROXY=https://gocenter.io

go mod download
go build kill-mysql-query.go