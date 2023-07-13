FROM golang:1.20-alpine AS builder

WORKDIR /usr/src/app/

COPY go.mod go.sum ./
ENV GOPROXY="https://goproxy.cn"
RUN go mod download
RUN go mod verify

COPY . ./

RUN GOOS=linux go build -ldflags="-w -s" -o /migrate ./cmd/goose/

CMD [ "/migrate", "up" ]
