# setup project and deps
FROM golang:1.25-bookworm AS init

WORKDIR /go/waffles/

COPY go.mod* go.sum* ./
RUN go mod download

COPY . ./

FROM init AS vet
RUN go vet ./...

# run tests
FROM init AS test
RUN go test -coverprofile c.out -v ./... && \
	echo "Statements missing coverage" && \
	grep -v -e " 1$" c.out

# build waffles binary
FROM init AS build
ARG LDFLAGS

RUN CGO_ENABLED=0 go build -ldflags="${LDFLAGS}" -o waffles

# build Go dependencies
FROM golang:1.25-bookworm AS build-go-deps

# Install Go tools
RUN CGO_ENABLED=0 go install github.com/toozej/wheresmyprompt@latest && \
    CGO_ENABLED=0 go install github.com/toozej/files2prompt@latest

# build Python dependencies
FROM python:3.13-bookworm AS build-python-deps

# Install uv and llm CLI tool
RUN pip install --no-cache-dir uv && \
    uv pip install --system --no-cache llm

# runtime image with Python support
FROM python:3.13-slim-bookworm

# Install runtime dependencies and create directories
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/* && \
    mkdir -p /go/bin /usr/local/bin

# Copy waffles binary
COPY --from=build /go/waffles/waffles /go/bin/waffles

# Copy Go tool binaries
COPY --from=build-go-deps /go/bin/wheresmyprompt /usr/local/bin/wheresmyprompt
COPY --from=build-go-deps /go/bin/files2prompt /usr/local/bin/files2prompt

# Copy Python environment with llm
COPY --from=build-python-deps /usr/local/lib/python3.13/site-packages /usr/local/lib/python3.13/site-packages
COPY --from=build-python-deps /usr/local/bin/llm /usr/local/bin/llm

# Set PATH to include our binaries
ENV PATH="/go/bin:/usr/local/bin:${PATH}"

# Set working directory
WORKDIR /workspace

# Run the binary
ENTRYPOINT ["/go/bin/waffles"]
