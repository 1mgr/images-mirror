# Use an official Golang runtime as a parent image
FROM golang:1.23-alpine

# Set the working directory in the container
WORKDIR /app

COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum
# Fetch dependencies
RUN go mod download

# Copy the current directory contents into the container at /app
COPY . /app

# Build the Go app
RUN go build -o main .

# Make port 80 available to the world outside this container
EXPOSE 8080

# Run the executable
CMD ["./main"]