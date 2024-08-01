FROM golang:1.22.5-alpine3.19 as builder

ADD . /go/src/app
WORKDIR /go/src/app

COPY go.mod ./
COPY go.sum ./
COPY *.go ./

RUN go mod download
RUN go mod tidy
RUN go build -o server

FROM scratch

COPY --from=builder /go/src/app/server /app/server

# Set the working directory
WORKDIR /app

# Command to run the binary
CMD ["./server"]
