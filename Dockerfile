FROM golang:1.24-alpine AS builder

WORKDIR /src

# Copy go.mod
COPY go.mod .
# Create an empty go.sum (will be populated by go mod download)
RUN touch go.sum

# Download dependencies with explicit package downloads
RUN go get github.com/gorilla/mux
RUN go get github.com/lib/pq
RUN go get github.com/google/uuid
RUN go mod tidy

# Copy the rest of the source code
COPY *.go .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/upload-service .

FROM alpine:3.21

RUN apk --no-cache --no-progress add ca-certificates tzdata \
    && update-ca-certificates \
    && rm -rf /var/cache/apk/*

RUN adduser \
    --disabled-password \
    --home /dev/null \
    --no-create-home \
    --shell /sbin/nologin \
    --gecos upload-service \
    --uid 10000 \
    upload-service

# Create uploads directory and set permissions
RUN mkdir -p /app/uploads && chown upload-service:upload-service /app/uploads

USER upload-service

COPY --from=builder /bin/upload-service /bin/

WORKDIR /bin

ENTRYPOINT ["/bin/upload-service"]
