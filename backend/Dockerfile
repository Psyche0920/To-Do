FROM golang:1.27.3-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /api ./cmd/api

FROM alpine:3.22
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 api
COPY --from=build /api /usr/local/bin/api
USER api
EXPOSE 8080
CMD ["api"]
