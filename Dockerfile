FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o terraform-gitlab-drift .

FROM alpine:3.23.3

RUN apk --no-cache add ca-certificates diffutils

COPY --from=builder /build/terraform-gitlab-drift /usr/local/bin/terraform-gitlab-drift

WORKDIR /workspace
