FROM golang:latest

# Make a standalone executable
ENV CGO_ENABLED=0

# Define the working directory
WORKDIR /usr/src/app

# Copy the Go module files (but not the source code); this enables the
# installation of dependencies to be cached and only rebuilt when the
# dependencies change
COPY go.mod go.sum ./

# Fetch the packages needed to build the application
RUN go mod download && go mod verify

# Copy the rest of the source code
COPY . .

# Now build the application
RUN go build -v -o certy


# Create the final container with only the application binary
FROM scratch

# Copy the binary
COPY --from=0 /usr/src/app/certy /usr/local/bin/

# Specify /data for data file storage
ENV DATA_DIR=/data

# Set the entrypoint
ENTRYPOINT ["certy"]
