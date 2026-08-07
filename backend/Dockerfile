FROM golang:1.25-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/yt-off ./cmd/server

FROM alpine:3.20

WORKDIR /app

ENV APP_HOST=0.0.0.0
ENV APP_PORT=8080

COPY --from=build /bin/yt-off /usr/local/bin/yt-off

EXPOSE 8080

CMD ["yt-off"]
