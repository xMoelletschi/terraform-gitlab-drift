FROM golang:1.25-alpine AS builder

ARG VERSION=dev

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s -X github.com/xMoelletschi/terraform-gitlab-drift/cmd.version=${VERSION}" -o terraform-gitlab-drift .

FROM alpine:3.23.3

RUN apk --no-cache add ca-certificates diffutils

COPY --from=builder /build/terraform-gitlab-drift /usr/local/bin/terraform-gitlab-drift

WORKDIR /workspace
