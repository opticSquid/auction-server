FROM golang:1.25-bookworm AS builder

WORKDIR /app

# Copy only necessary files
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o auction-server .

# Create a new stage for the final image
FROM scratch

# Copy the application.yaml from resources directory from the build context to final image
COPY --from=builder /app/resources/application.yaml /resources/application.yaml
# Copy the built binary from the previous stage
COPY --from=builder /app/auction-server /

# Set the working directory
WORKDIR /

# Run binary
CMD ["./auction-server"]
