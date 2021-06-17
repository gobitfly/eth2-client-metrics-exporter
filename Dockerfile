FROM golang:1.16 as builder
ADD . /app
WORKDIR /app
RUN make

FROM ubuntu:18.04
RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/bin/eth2-client-metrics-exporter /usr/local/bin/eth2-client-metrics-exporter
ENTRYPOINT ["/usr/local/bin/eth2-client-metrics-exporter"]
