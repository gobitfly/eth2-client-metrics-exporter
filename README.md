# eth2-client-metrics-exporter

A sidecar for exporting [eth2-client-metrics](https://github.com/gobitfly/eth2-client-metrics).

## usage

First start `beaconnode` and `validator` (with metrics-endpoints enabled), then start the `eth2-client-metrics-exporter` and point it to your beaconnode and validator:

```bash
make
./bin/eth2-client-metrics-exporter-linux-amd64 \
    --server.address='https://prater.beaconcha.in/api/v1/client/metrics?apikey=<beaconcha.in-apikey>&machine=<machine-name>' \
    --beaconnode.type=prysm \
    --beaconnode.address=http://localhost:9090 \
    --validator.type=prysm \
    --validator.address=http://localhost:9091
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
