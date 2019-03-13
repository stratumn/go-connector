FROM golang:1.12-stretch AS build-env

# All these steps will be cached
RUN mkdir /home/connector
WORKDIR /home/connector
COPY go.mod .
COPY go.sum .

# Get dependencies - will also be cached if we won't change mod/sum
RUN go mod download

# COPY the source code as the last step
COPY . .

# Build the server
RUN go build -o connector

FROM scratch

# Copy binaries and plugins
COPY --from=build-env /home/connector/validation.so /go/bin/validation.so
COPY --from=build-env /home/connector/api.so /go/bin/api.so
COPY --from=build-env /home/connector/decryption.so /go/bin/decryption.so

COPY --from=build-env /home/connector/connector /go/bin/connector

ENTRYPOINT ["/go/bin/connector"]
