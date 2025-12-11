FROM golang:1.23-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/run

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /app/server /app/server

EXPOSE 8080

USER 65532:65532
ENTRYPOINT ["/app/server"]
