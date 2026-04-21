FROM golang:1.20-alpine AS build
WORKDIR /src
RUN apk add --no-cache ca-certificates
COPY go.mod ./
RUN go mod download
COPY . .
ARG SERVICE=sample-service
ARG BUILD_ID=local
ARG BUILD_TAG=dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -trimpath \
  -ldflags="-s -w -X=example.com/sample-service/internal/buildinfo.BuildID=${BUILD_ID} -X=example.com/sample-service/internal/buildinfo.BuildTag=${BUILD_TAG}" \
  -o /out/server ./cmd/${SERVICE}

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/server /server
USER 65534:65534
ENV PORT=8080
ENTRYPOINT ["/server"]
