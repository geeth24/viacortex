FROM golang:1.22

WORKDIR /app

# Install air
RUN go install github.com/cosmtrek/air@v1.49.0

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Expose both HTTP and HTTPS ports
EXPOSE 80
EXPOSE 443
EXPOSE 8080

CMD ["/go/bin/air"]