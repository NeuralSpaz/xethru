#!/usr/bin/make -f

SHELL=/bin/bash

all:
	GOARCH=amd64 GOOS=linux go build -o xethru-respiration-linux-64 .
	gzip --best xethru-respiration-linux-64
	GOARCH=amd64 GOOS=windows go build -o xethru-respiration-windows-64.exe .
	gzip --best xethru-respiration-windows-64.exe
	GOARCH=amd64 GOOS=darwin go build -o xethru-respiration-mac-64 .
	gzip --best xethru-respiration-mac-64
	GOARCH=386 GOOS=linux go build -o xethru-respiration-linux-32 .
	gzip --best xethru-respiration-linux-32
	GOARCH=386 GOOS=windows go build -o xethru-respiration-windows-32.exe .
	gzip --best xethru-respiration-windows-32.exe


clean:
	rm xethru-respiration-*
