# Use minimal Go image
FROM golang:1.23-alpine

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first (for caching layer)
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy the rest of your code
COPY . .

# Build binary
RUN go build -o colly-scrapper-op

# Expose app port
EXPOSE 3000

# Run the app
CMD ["./colly-scrapper-op"]
