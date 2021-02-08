# mTLS Proxy

## Generate Certificate

```sh
CAROOT=. mkcert -ecdsa server
CAROOT=. mkcert -ecdsa -client client
```

## Build

```sh
go build -o mtls-proxy
```

## Run

### Server

```sh
mtls-proxy \
  -mode=server \
  -ca=rootCA.pem -cert=server.pem -key=server-key.pem \
  -server-addr=localhost:9000 \
  -addr=:3000
```

### Client

```sh
mtls-proxy \
  -mode=client \
  -ca=rootCA.pem -cert=client-client.pem -key=client-client-key.pem \
  -server-addr=localhost:3000 -server-name=server \
  -addr=:4000
```
