FROM golang:1.10
MAINTAINER Sangyun Han <sangyun628@gmail.com>

# setup golang
RUN apt-get update
RUN apt-get install vim -y
RUN go get github.com/sangyun-han/monitor4sdn

# Ports
# 6653 - OpenFlow
# 8086 - InfluxDB
EXPOSE 6653 8086

WORKDIR /go/src/github.com/sangyun-han/monitor4sdn
ENTRYPOINT  ["go", "run", "main.go"]