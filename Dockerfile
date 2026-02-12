FROM golang:1.24-alpine AS build
WORKDIR /src
ENV GOPROXY=http://host.docker.internal:3000|direct
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /noknok ./cmd/noknok

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=build /noknok /usr/local/bin/noknok
ENTRYPOINT ["noknok"]
