# nocors

It's a very simple HTTP reverse proxy that disables CORS in the upstream
server.
It intercepts the preflight requests and adds the proper "Access-Control-Allow-Origin" header
to all responses.

usage:

```bash
nocors listen_address dest_address
```

example:

```bash
nocors 127.0.0.1:8081 127.0.0.1:8080
```
