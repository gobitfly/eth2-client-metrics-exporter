# eth2-client-metrics-exporter

a sidecar for exporting [eth2-client-metrics](https://github.com/gobitfly/eth2-client-metrics)

## usage

first start `beaconnode` and `validator` (with metrics-endpoints enabled), then start the `eth2-client-metrics-exporter` and point it to your beaconnode and validator:

```bash
# network: prater, client: prysm, 
docker run gobitfly/eth2-client-metrics \
    --server.address=https://prater.beaconcha.in/api/v1/client/metrics?apikey=<beaconcha.in-apikey>&machine=<machine-name> \
    --beaconnode.type=prysm \
    --beaconnode.address=http://localhost:9090 \
    --validator.type=prysm \
    --validator.address=http://localhost:9091
```

## info


