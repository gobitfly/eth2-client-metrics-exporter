# eth2-client-metrics-exporter

A sidecar for exporting [eth2-client-metrics](https://github.com/gobitfly/eth2-client-metrics).

## How to use

1. Start `beaconnode` and/or `validator` with metrics-endpoints enabled.

2. Get your `server.address` by signing in to https://beaconcha.in/user/settings#app and copy the URL.

3. Start the eth2-client-metrics-exporter and point it to your `beaconnode` and/or `validator`.

*For a more comprehensive guide please take a look at the [beaconchain-knowledge-base](https://kb.beaconcha.in/beaconcha.in-explorer/mobile-app-less-than-greater-than-beacon-node).*

### Example with binary

Make sure to run your beaconnode and/or validator with metrics enabled (in case of Nimbus for example use these flags `--metrics --metric-port=8008`).

Then point the exporter at this metrics-endpoint:

```bash
./eth2-client-metrics-exporter-linux-amd64 \
    --server.address='https://beaconcha.in/api/v1/client/metrics?apikey=<beaconcha.in-apikey>&machine=<machine-name>' \
    --beaconnode.type=nimbus \
    --beaconnode.address=http://localhost:8008/metrics \
```

### Example with docker

```yaml
version: "3.7"
services:
  # run prysm-node with metrics enabled
  prysm-node:
    image: gcr.io/prysmaticlabs/prysm/beacon-chain
    command:
      - --accept-terms-of-use
      - --datadir=/data
      - --monitoring-port=9090
      - --monitoring-host=0.0.0.0
    volumes:
      - ./docker-volumes/prysm-node:/data
    restart: always

  # run prysm-validator with metrics enabled
  prysm-validator:
    image: gcr.io/prysmaticlabs/prysm/validator
    restart: unless-stopped
    command:
      - --accept-terms-of-use
      - --beacon-rpc-provider=prysm-node:4000
      - --wallet-password-file=/v/wallet-password.txt
      - --monitoring-port=9090
      - --monitoring-host=0.0.0.0
    volumes:
      - ./docker-volumes/prysm-validator:/home/.eth2
      - ./wallet-password.txt:/v/wallet-password.txt

  # point exporter to metrics of beaconnode and/or validator
  eth2-client-metrics-exporter:
    image: gobitfly/eth2-client-metrics-exporter
    restart: unless-stopped
    command:
      - --server.address=https://beaconcha.in/api/v1/client/metrics?apikey=<apikey>&machine=<machine>
      - --system.partition=/host/rootfs
      - --beaconnode.type=prysm
      - --beaconnode.address=http://prysm-node:9090/metrics
      - --validator.type=prysm
      - --validator.address=http://prysm-validator:9090/metrics
    volumes:
      - /sys:/host/sys:ro
      - /proc:/host/proc:ro
      - /:/host/rootfs:ro
    environment:
      - HOST_PROC=/host/proc
      - HOST_SYS=/host/sys
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
* Teku
  * Teku has its own metrics-exporter built in and it works out of the box, no need for this extra program right now
* Lodestar
  * has its own metrics-exporter built-in which can be enabled by passing the `--monitoring.endpoint` CLI flag, see [client monitoring](https://chainsafe.github.io/lodestar/usage/client-monitoring/) for details.
* Nimbus
  * beaconnode is partially implemented
  * validator is not implemented
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
