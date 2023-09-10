FROM golang:latest
WORKDIR /app
COPY . .
RUN go build -o target cmd/target/main.go
RUN chmod +x target

