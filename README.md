# eth2-client-metrics-exporter

A sidecar for exporting [eth2-client-metrics](https://github.com/gobitfly/eth2-client-metrics).

## How to use

1. Get your server.address by signing in to https://beaconcha.in/user/settings#app and copy the URL.

2. Download this tool from our releases page or compile from source (see below). For Raspberry PI's or other ARM based CPU's, use `eth2-client-metrics-exporter-linux-arm64`. Otherwise `eth2-client-metrics-exporter-linux-amd64`.

2. Start `beaconnode` and `validator` (with metrics-endpoints enabled), then start the `eth2-client-metrics-exporter` and point it to your beaconnode and validator:

### Prysm

Replace the server.address URL with the one retrieved from Step 1.

#### Node & Validator

```bash
./eth2-client-metrics-exporter-linux-amd64 \
    --server.address='https://beaconcha.in/api/v1/client/metrics?apikey=<beaconcha.in-apikey>&machine=<machine-name>' \
    --beaconnode.type=prysm \
    --beaconnode.address=http://localhost:8080/metrics \
    --validator.type=prysm \
    --validator.address=http://localhost:8081/metrics
```

If you want to monitor only the node or only the validator, ommit either the beaconnode or validator flags.  
  

### Nimbus

- Make sure you started your Nimbus node with
```
--metrics --metric-port=8008
```

Replace the server.address URL with the one retrieved from Step 1. Then run the metrics exporter with:

```bash
./eth2-client-metrics-exporter-linux-amd64 \
    --server.address='https://beaconcha.in/api/v1/client/metrics?apikey=<beaconcha.in-apikey>&machine=<machine-name>' \
    --beaconnode.type=nimbus \
    --beaconnode.address=http://localhost:8008/metrics \
```
  
  

## Build
- Requirement: Go 1.16

```bash
git clone https://github.com/gobitfly/eth2-client-metrics-exporter.git
cd eth2-client-metrics-exporter 
make
```

## client support status

* Lighthouse
  * Lighthouse has its own metrics-exporter built in and it works out of the box, no need for this extra program right now
* Nimbus
  * partial support
* Prysm
  * missing
    * beaconnode
      * `sync_eth2_fallback_configured`
      * `sync_eth2_fallback_connected`
      * `network_libp2p_bytes_total_transmit`
      * `sync_eth1_connected`
      * `sync_eth2_synced`
      * `sync_eth1_fallback_configured`
      * `sync_eth1_fallback_connected`
      * `slasher_active`
    * validator
      * `sync_eth2_fallback_configured`
      * `sync_eth2_fallback_connected`
* Teku
  * not implemented yet
