# syntax=docker/dockerfile:1

FROM golang:1.24 AS go-build-stage
WORKDIR /app
COPY ./go.mod ./go.sum ./
RUN go mod download
ADD . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./src/cmd/main.go

FROM gcr.io/distroless/base-debian12 AS build-release-stage
WORKDIR /
COPY --from=go-build-stage /api /api
ENTRYPOINT ["/api"]

