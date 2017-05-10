# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.8

# Copy the local package files to the container's workspace.
ADD sensor-exporter /go/src/github.com/ncabatoff/sensor-exporter

# Build the outyet command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN apt-get update
RUN apt-get --yes install libsensors4-dev
RUN go get github.com/ncabatoff/gosensors github.com/prometheus/client_golang/prometheus && go install github.com/ncabatoff/sensor-exporter

# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/sensor-exporter

# Document that the service listens on port 9255.
EXPOSE 9255
