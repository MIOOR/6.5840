#!/usr/bin/env bash
go build -race -buildmode=plugin -gcflags="all=-N -l"  ../mrapps/wc.go
rm mr-out*
          
