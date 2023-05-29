# Build the manager binary
FROM golang:1.17 as builder


WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o ca_generator main.go


FROM alpine:3.6

ARG NONROOT_UID=65532
ARG NONROOT_GID=65532

WORKDIR /ca-generator
COPY --from=builder /workspace/ca_generator .

ADD https://storage.googleapis.com/kubernetes-release/release/v1.20.5/bin/linux/amd64/kubectl /usr/local/bin/kubectl

ADD https://dl.k8s.io/release/v1.26.0/bin/linux/amd64/kubectl /usr/local/bin/kubectl

RUN chmod +x /usr/local/bin/kubectl \
    && adduser -u $NONROOT_UID -D nonroot $NONROOT_GID \
    && chown nonroot:root /ca-generator

USER nonroot
