FROM golang:1.24 AS builder
WORKDIR /app
COPY . .
RUN go build -o ta .

FROM oven/bun:latest
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /app/ta /usr/local/bin/ta
COPY ta.toml .htmlvalidate.json .stylelintrc.json eslint.config.js package.json bun.lock ./
RUN bun install
ENTRYPOINT ["ta"]
