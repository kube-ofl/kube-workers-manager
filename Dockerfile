FROM golang:1.22.5-alpine3.19 as base

FROM base as dev

ADD . /go/src/app
WORKDIR /go/src/app

COPY go.mod ./
COPY go.sum ./
COPY *.go ./

RUN go mod download
RUN go mod tidy
RUN go build -o server

CMD [ "./server" ]