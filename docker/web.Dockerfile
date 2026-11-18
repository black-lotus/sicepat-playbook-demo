# Image development untuk web. Source di-mount dari host lewat docker-compose.yml.
ARG NODE_VERSION
FROM node:${NODE_VERSION}-alpine

WORKDIR /app
