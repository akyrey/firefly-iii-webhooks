# syntax=docker/dockerfile:1

## STEP 1 - BUILD
FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

# Install go modules
RUN go mod download

COPY ./ ./

# Compile the app into /ff-webhooks directory
RUN CGO_ENABLED=0 GOOS=linux go build -o /ff-webhooks ./cmd/firefly-iii-webhooks

## STEP 2 - DEPLOY
FROM alpine:3.16

WORKDIR /

COPY --from=build /ff-webhooks /ff-webhooks
COPY --from=build /etc/passwd /etc/passwd

USER guest

EXPOSE 8080

ENTRYPOINT [ "/ff-webhooks", "-addr", ":8080" ]
