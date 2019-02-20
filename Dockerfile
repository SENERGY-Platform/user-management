FROM golang:1.11


COPY . /go/src/user-management
WORKDIR /go/src/user-management

ENV GO111MODULE=on

RUN go build

EXPOSE 8080

CMD ./user-management