#!/bin/bash

echo "Downloading beaconcha.in metric exporter..."
curl -s https://api.github.com/repos/gobitfly/eth2-client-metrics-exporter/releases/latest | grep 'eth2-client-metrics-exporter-linux-amd64' | cut -d : -f 2,3 |  tr -d \" | wget -qi -
chmod +x eth2-client-metrics-exporter-linux-amd64
echo "Done."
