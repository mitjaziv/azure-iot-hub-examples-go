#!/usr/bin/env bash
go build -o build/sas sas-token/sas-token.go
go build -o build/sac sa-cert/sa-cert.go
go build -o build/cac ca-cert/ca-cert.go
