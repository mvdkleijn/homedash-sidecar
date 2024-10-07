################
# Build binary #
################
FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.22 AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH}

WORKDIR /build

COPY . .

RUN go mod download
RUN go vet -v
RUN go test -v
RUN go build -ldflags="-w -s" -o app .

#####################
# Build final image #
#####################
FROM --platform=${TARGETPLATFORM:-linux/amd64} gcr.io/distroless/static-debian11:nonroot

COPY --from=builder /build/app /

ENTRYPOINT ["/app"]