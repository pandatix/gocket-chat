# Build stage
FROM golang:1.17.0 as builder
WORKDIR /go/src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0
RUN go build -o /go/bin/gocket-chat cmd/gocket-chat/main.go

# Prod stage
FROM scratch
COPY --from=builder /go/bin/gocket-chat /gocket-chat
ENTRYPOINT [ "/gocket-chat" ]
