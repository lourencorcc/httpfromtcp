# httpfromtcp

A small Go project that rebuilds HTTP on top of raw TCP for fun and learning.

## What it does/will do

- [x] Reads an HTTP request, directly from a TCP connection and effectively handles chunks
- [x] Parses request line, headers and body correctly following the RFC as much as possible (it's a long read ok.)
- Handles requests and sends responses back.

## Why

Created this repo and project to get hands-on with Go and networks by reimplementing HTTP behavior from TCP only.

## AI Usage

- 0 bugs were fixed using AI
- 0 lines of code were written using AI
- AI was used for educational purposes and curiosity 

## Status
- Currently working on writing a server interface to use the request parser implemented
