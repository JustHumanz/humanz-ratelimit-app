FROM golang:latest as builder
COPY app .
RUN CGO_ENABLED=0 GOOS=linux go build -o server
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 2525
CMD ["./server"]