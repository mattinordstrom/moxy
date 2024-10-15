FROM golang:1.22-alpine

WORKDIR /moxyapp

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 9097

# Command to run the Go application directly
CMD ["go", "run", "main.go"]
