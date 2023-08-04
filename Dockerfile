# syntax=docker/dockerfile:1

FROM node:14.16 AS frontend-builder

WORKDIR /app

COPY frontend/package.json ./
COPY frontend/package-lock.json ./
RUN npm install

COPY frontend/ ./
RUN npm run build

FROM golang:1.19 AS backend-builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
USER 0
RUN go mod download

COPY pkg/ pkg/
COPY cmd/ cmd/
RUN CGO_ENABLED=0 go build -o backend cmd/backend/backend.go

FROM alpine:3.15
WORKDIR /
RUN ls -lsaR
COPY --from=backend-builder app/backend /
COPY --from=frontend-builder app/dist/ /frontend/build/
