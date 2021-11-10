FROM golang:latest as buildenv

WORKDIR /build

COPY go.mod go.sum ./

# Download all the dependencies
RUN go mod download

COPY . .

WORKDIR /build/manager

RUN GOOS=linux go build -a -installsuffix cgo -o manager .

FROM ubuntu
MAINTAINER Andrii Kondratiuk <andriikondratiukv@gmail.com>
WORKDIR /app
COPY --from=buildenv /build/manager/manager .

ENTRYPOINT ["./manager"]