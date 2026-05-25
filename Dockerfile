FROM golang:1.23-alpine3.20 AS builder

ENV CGO_ENABLED=0
ENV TZ Asia/Kolkata

# Creates an app directory to hold your app’s source code
WORKDIR /annualstatement

COPY go.mod .
COPY go.sum .
COPY . .

# Builds your app with optional configuration
RUN go build -o /annualstatement/bin .

ENTRYPOINT ["/annualstatement/bin"]
