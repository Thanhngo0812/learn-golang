FROM golang:1.25.6-alpine

WORKDIR /app

# Copy các file quản lý thư viện trước để tận dụng cache
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build ứng dụng
RUN go build -o main .

EXPOSE 8080

CMD ["./main"]