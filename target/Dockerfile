FROM golang:latest as buildenv

WORKDIR /build

COPY go.mod go.sum ./

# Download all the dependencies
RUN go mod download

COPY . .

WORKDIR /build/sites/target

RUN GOOS=linux go build -a -installsuffix cgo -o target .

FROM ubuntu
MAINTAINER Andrii Kondratiuk <andriikondratiukv@gmail.com>
WORKDIR /app
COPY --from=buildenv /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=buildenv /build/sites/target/target .

ENTRYPOINT ["./target"]