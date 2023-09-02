# nocors

A simple HTTP reverse proxy that disables CORS in the upstream server.

It intercepts all the preflight requests and adds the proper
"Access-Control-Allow-Origin" header to all responses.

## Install

```bash
go install github.com/manelmontilla/nocors/cmd/nocors@latest
```

## Usage

```bash
nocors listen_address dest_address
```

## Example

The following example disables CORS in the upstream server listening in the
address: 127.0.0.1:8080

```bash
nocors 127.0.0.1:8081 127.0.0.1:8080
```

To access the server listening in the port 8080 with CORS disabled, just use
127.0.0.1:8081 as base address in your browser.
