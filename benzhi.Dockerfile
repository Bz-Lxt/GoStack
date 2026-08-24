FROM golang:1.25.12
WORKDIR /app
COPY backend/ /app/
ENV GOTOOLCHAIN=local TZ=Asia/Shanghai
RUN go test ./...
