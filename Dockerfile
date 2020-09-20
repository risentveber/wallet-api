FROM golang:1.15-alpine as builder

WORKDIR /app/
COPY . .
RUN apk add -U --no-cache ca-certificates
ENV CGO_ENABLED=0
ENV GOOS=linux
RUN env
RUN go build -o /app/cmd/api/api.bin /app/cmd/api

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/cmd/api/api.bin /bin/api
COPY ./migrations /migrations

EXPOSE 8080
ENTRYPOINT ["/bin/api"]