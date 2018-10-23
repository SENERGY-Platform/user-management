FROM golang:1.11


COPY . /go/src/users
WORKDIR /go/src/users

ENV GO111MODULE=on

RUN go build

EXPOSE 8080

CMD ./users