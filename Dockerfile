FROM golang:1.15.8

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0
RUN mkdir -p /workspace
WORKDIR /workspace
ADD go.mod go.sum ./
RUN go mod download
ADD . .
RUN go build -trimpath -o mtls-proxy -ldflags "-w -s" .

# ---------------------------------------------------------------------------------

FROM scratch

COPY --from=0 /workspace/mtls-proxy /
ENTRYPOINT ["/mtls-proxy"]
