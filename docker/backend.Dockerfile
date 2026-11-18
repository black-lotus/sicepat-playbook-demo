# Image development untuk backend. Source di-mount dari host lewat docker-compose.yml.
ARG GO_VERSION
FROM golang:${GO_VERSION}-alpine

WORKDIR /app
