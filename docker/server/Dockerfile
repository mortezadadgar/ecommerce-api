FROM golang:1.20-alpine AS builder

WORKDIR /usr/src/app/

COPY go.mod /go.sum ./
ENV GOPROXY="https://goproxy.cn"
RUN go mod download
RUN go mod verify

ENV USER="user"
RUN adduser --disabled-password  \
			--shell "/sbin/nologin" \
			"$USER"

COPY . ./

RUN GOOS=linux go build -ldflags="-w -s" -o /server ./cmd/ecommerce/

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

USER $USER:$USER

COPY --from=builder /server /home/$USER/

WORKDIR /home/$USER

ENTRYPOINT [ "./server" ]
