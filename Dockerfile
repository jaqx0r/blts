FROM docker.io/golang:alpine AS builder
WORKDIR /blts
RUN apk add --update make
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN make clean all

FROM docker.io/alpine AS servers
COPY servers.sh servers.sh
COPY replace.sh replace.sh
COPY --from=builder /blts/s s
COPY --from=builder /blts/lb lb
EXPOSE 8000 8001 8002 8003 8004 8005 8006 8007 8008 8009 9001
ENV zipkin "--zipkin="
ENTRYPOINT ./servers.sh ${zipkin}
