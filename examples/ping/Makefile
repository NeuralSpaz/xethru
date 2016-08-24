#!/usr/bin/make -f

SHELL=/bin/bash

all:
	GOARCH=amd64 GOOS=linux go build -o xethru-ping-linux-64 .
	gzip --best xethru-ping-linux-64
	GOARCH=amd64 GOOS=windows go build -o xethru-ping-windows-64.exe .
	gzip --best xethru-ping-windows-64.exe
	GOARCH=amd64 GOOS=darwin go build -o xethru-ping-mac-64 .
	gzip --best xethru-ping-mac-64
	GOARCH=386 GOOS=linux go build -o xethru-ping-linux-32 .
	gzip --best xethru-ping-linux-32
	GOARCH=386 GOOS=windows go build -o xethru-ping-windows-32.exe .
	gzip --best xethru-ping-windows-32.exe


clean:
	rm xethru-ping-*
