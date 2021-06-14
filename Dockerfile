FROM golang:alpine AS builder
COPY . /app
WORKDIR /app
RUN ["go", "build","-o","gitserve","."]

FROM alpine
WORKDIR /app
COPY --from=builder /app/gitserve .
ENTRYPOINT ["./gitserve"]