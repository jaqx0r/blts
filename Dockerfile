FROM golang:alpine AS builder
RUN apk add --update make
WORKDIR /blts
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN make

FROM alpine AS servers
COPY servers.sh servers.sh
COPY --from=builder /blts/s/s s/s
COPY --from=builder /blts/lb/lb lb/lb
EXPOSE 8000 8001 8002 8003 8004 8005 8006 8007 8008 8009 9001
ARG zipkin
ENTRYPOINT ["./servers.sh"]