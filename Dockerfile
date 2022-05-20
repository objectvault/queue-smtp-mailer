## APP BUILD ENVIRONMENT ##
FROM golang:1.18-alpine as builder

# Set Working Directory
WORKDIR /usr/src/app

# Add GIT Package to Build Environment
RUN apk --no-cache --no-progress add \
    git

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy Application Source
COPY . .

# Build Application to /usr/local/bin/app
RUN go build -v -o /usr/local/bin/app .

## APP RUN ENVIRONMENT ##
FROM alpine:latest

# SET Working Directory
WORKDIR /app

# Copy License and README File
COPY  LICENSE.md README.md ./

# Copy Application from Build Environment
COPY --from=builder /usr/local/bin/app /app/mailer

# Set Server Configuration File
# copy example/docker.local.json server.json
# use docker run ... -v /path/conf.json:/app/server.json:ro ...

# PORTS
EXPOSE 3000

# Execute Command
CMD ["./mailer", "-c", "./mailer.json"]
