package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"runtime"
	"sync"
	"time"

	promModel "github.com/prometheus/client_model/go"
	promExpfmt "github.com/prometheus/common/expfmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/sirupsen/logrus"
)

// Build information. Populated at build-time
var (
	Version      = "undefined"
	GitDate      = "undefined"
	GitDateShort = "undefined"
	GitCommit    = "undefined"
	BuildDate    = "undefined"
	GoVersion    = runtime.Version()
)

var options = struct {
	ServerAddress     string
	ServerTimeout     time.Duration
	BeaconnodeType    string
	BeaconnodeAddress string
	ValidatorType     string
	ValidatorAddress  string
	Interval          time.Duration
	Partition         string
	Debug             bool
}{}

type ClientType string

const (
	PrysmBeaconnodeMetricsClientType  ClientType = "prysm-beaconnode-metrics"
	PrysmValidatorMetricsClientType   ClientType = "prysm-validator-metrics"
	NimbusBeaconnodeMetricsClientType ClientType = "nimbus-beaconnode-metrics"
)

type ClientEndpoint struct {
	Type    ClientType
	Address string
}

type ServerResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

var clientEndpoints = []ClientEndpoint{}

var httpClient *http.Client

var specVersion = int64(2)
var exporterVersion = ""

func main() {
	flag.BoolVar(&options.Debug, "debug", false, "enable debugging")
	flag.DurationVar(&options.Interval, "interval", time.Second*62, "interval of sending metrics to server")
	flag.StringVar(&options.ServerAddress, "server.address", "", "address of server to push metrics to")
	flag.DurationVar(&options.ServerTimeout, "server.timeout", time.Second*10, "timeout for sending data to the server")
	flag.StringVar(&options.Partition, "system.partition", "/", "mountpoint of partition which will be tracked for usage, if empty-string the highest usage of any partition will be recorded")
	flag.StringVar(&options.BeaconnodeType, "beaconnode.type", "prysm", "endpoint to scrape metrics from")
	flag.StringVar(&options.BeaconnodeAddress, "beaconnode.address", "", "address of beaconnode-endpoint to scrape metrics from (eg: http://localhost:9090/metrics), disabled if empty string")
	flag.StringVar(&options.ValidatorType, "validator.type", "prysm", "endpoint to scrape metrics from")
	flag.StringVar(&options.ValidatorAddress, "validator.address", "", "address of validator-endpoint to scrape metrics from (eg: http://localhost:9090/metrics), disabled if emtpy string")
	versionFlag := flag.Bool("version", false, "show version and exit")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("%v\n", Version)
		return
	}

	if options.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if options.ServerAddress == "" {
		logrus.Fatal("Server address not provided.")
	}

	if options.BeaconnodeAddress != "" {
		var clientType ClientType
		switch options.BeaconnodeType {
		case "prysm":
			clientType = PrysmBeaconnodeMetricsClientType
		case "nimbus":
			clientType = NimbusBeaconnodeMetricsClientType
		default:
			logrus.Fatal("invalid beaconnode.type")
		}
		clientEndpoints = append(clientEndpoints, ClientEndpoint{
			Type:    clientType,
			Address: options.BeaconnodeAddress,
		})
	}

	if options.ValidatorAddress != "" {
		var clientType ClientType
		switch options.BeaconnodeType {
		case "prysm":
			clientType = PrysmValidatorMetricsClientType
		default:
			logrus.Fatal("invalid beaconnode.type")
		}
		clientEndpoints = append(clientEndpoints, ClientEndpoint{
			Type:    clientType,
			Address: options.ValidatorAddress,
		})
	}

	if options.BeaconnodeAddress == "" && options.ValidatorAddress == "" {
		logrus.Fatal("Neither beacon node nor validator address provided.")
	}

	exporterVersion = fmt.Sprintf("beaconcha.in@%v", GitCommit)

	httpClient = &http.Client{
		Timeout: options.ServerTimeout,
	}

	logrus.WithFields(logrus.Fields{
		// "ServerAddress": options.ServerAddress, // may contain secrets, don't log
		"ServerTimeout":     options.ServerTimeout,
		"BeaconnodeType":    options.BeaconnodeType,
		"BeaconnodeAddress": options.BeaconnodeAddress,
		"ValidatorType":     options.ValidatorType,
		"ValidatorAddress":  options.ValidatorAddress,
		"Interval":          options.Interval,
		"Partition":         options.Partition,
		"Debug":             options.Debug,
		"Version":           exporterVersion,
	}).Infof("starting exporter")

	collectDataLoop()
}

func collectDataLoop() {
	t := time.NewTicker(options.Interval)
	defer t.Stop()
	for {
		t0 := time.Now()
		d, err := collectData()
		if err != nil {
			logrus.WithError(err).Error("failed collecting data")
			time.Sleep(time.Second * 10)
			t.Reset(options.Interval)
			continue
		}
		logrus.WithFields(logrus.Fields{"duration": time.Since(t0)}).Info("collected data")

		t1 := time.Now()
		err = sendData(d)
		if err != nil {
			logrus.WithError(err).Error("failed sending data")
			time.Sleep(time.Second * 10)
			t.Reset(options.Interval)
			continue
		}
		logrus.WithFields(logrus.Fields{"duration": time.Since(t1)}).Info("sent data")

		select {
		case <-t.C:
		}
	}
}

func collectData() ([]interface{}, error) {
	var wg sync.WaitGroup

	results := make(chan interface{}, len(clientEndpoints)+1)
	ts := uint64(time.Now().UnixNano() / int64(time.Millisecond))

	wg.Add(1)
	go func() {
		defer wg.Done()
		d, err := getSystemData(ts)
		if err != nil {
			logrus.WithFields(logrus.Fields{"error": err, "type": "system"}).Errorf("failed getting data")
			return
		}
		results <- d
	}()

	for _, c := range clientEndpoints {
		wg.Add(1)
		go func(c ClientEndpoint) {
			defer wg.Done()
			switch c.Type {
			case PrysmBeaconnodeMetricsClientType:
				d, err := getPrysmBeaconnodeData(c.Address, ts)
				if err != nil {
					logrus.WithFields(logrus.Fields{"error": err, "address": c.Address, "type": c.Type}).Errorf("failed getting data")
					return
				}
				results <- d
			case PrysmValidatorMetricsClientType:
				d, err := getPrysmValidatorData(c.Address, ts)
				if err != nil {
					logrus.WithFields(logrus.Fields{"error": err, "address": c.Address, "type": c.Type}).Errorf("failed getting data")
					return
				}
				results <- d
			case NimbusBeaconnodeMetricsClientType:
				d, err := getNimbusBeaconnodeData(c.Address, ts)
				if err != nil {
					logrus.WithFields(logrus.Fields{"error": err, "address": c.Address, "type": c.Type}).Errorf("failed getting data")
					return
				}
				results <- d
			default:
				logrus.Fatalf("unknown client-endpoint-type: %v", c.Type)
			}
			return
		}(c)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	result := []interface{}{}
	for r := range results {
		result = append(result, r)
	}
	return result, nil
}

func sendData(data []interface{}) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		err = fmt.Errorf("failed marshaling data: %w", err)
		return err
	}

	logrus.WithFields(logrus.Fields{"json": fmt.Sprintf("%s", dataJSON)}).Debug("sending data")

	req, err := http.NewRequest("POST", options.ServerAddress, bytes.NewBuffer(dataJSON))
	if err != nil {
		err = fmt.Errorf("failed creating request: %w", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf("failed sending request: %w", err)
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK || len(body) != 0 {
		err = fmt.Errorf("got error-response from server: %s", body)
		return err
	}
	return nil
}

func getSystemData(ts uint64) (*SystemData, error) {
	systemData := &SystemData{}

	systemData.Version = specVersion
	systemData.Timestamp = ts
	systemData.ExporterVersion = exporterVersion
	systemData.Process = "system"

	cpuThreads, err := cpu.Counts(true)
	if err != nil {
		return nil, fmt.Errorf("failed getting cpu_threads: %w", err)
	}
	systemData.CPUThreads = int64(cpuThreads)

	cpuCores, err := cpu.Counts(false)
	if err != nil {
		return nil, fmt.Errorf("failed getting cpu_cores: %w", err)
	}
	systemData.CPUCores = int64(cpuCores)

	cpuTimes, err := cpu.Times(false)
	if err != nil {
		return nil, fmt.Errorf("failed getting cpu times: %w", err)
	}
	for _, t := range cpuTimes {
		systemData.CPUNodeIdleSecondsTotal += uint64(t.Idle)
		systemData.CPUNodeUserSecondsTotal += uint64(t.User)
		systemData.CPUNodeIOWaitSecondsTotal += uint64(t.Iowait)
		// note: currently beaconcha.in expects this to be everything
		systemData.CPUNodeSystemSecondsTotal += uint64(t.System) + uint64(t.Iowait) + uint64(t.User) + uint64(t.Idle)
	}

	memStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed getting memory stats: %w", err)
	}
	systemData.MemoryNodeBytesTotal = memStat.Total
	systemData.MemoryNodeBytesFree = memStat.Free
	systemData.MemoryNodeBytesCached = memStat.Cached
	systemData.MemoryNodeBytesBuffers = memStat.Buffers

	if options.Partition != "" {
		stat, err := disk.Usage(options.Partition)
		if err != nil {
			return nil, fmt.Errorf("failed getting disk partition stats for mountpoint: %s: %w", options.Partition, err)
		}
		systemData.DiskNodeBytesTotal += stat.Total
		systemData.DiskNodeBytesFree += stat.Free
	} else {
		parts, err := disk.Partitions(true)
		if err != nil {
			return nil, fmt.Errorf("failed getting disk partitions: %w", err)
		}
		var mostUsedPartStat *disk.UsageStat
		for _, p := range parts {
			stat, err := disk.Usage(p.Mountpoint)
			if err != nil {
				logrus.WithFields(logrus.Fields{"error": err, "mountpoint": p.Mountpoint}).Error("failed getting disk partition stats")
			} else {
				if mostUsedPartStat == nil || stat.UsedPercent > mostUsedPartStat.UsedPercent {
					mostUsedPartStat = stat
					systemData.DiskNodeBytesTotal += stat.Total
					systemData.DiskNodeBytesFree += stat.Free
				}
			}
		}
		logrus.WithFields(logrus.Fields{"path": mostUsedPartStat.Path, "usedPercent": mostUsedPartStat.UsedPercent, "totalBytes": mostUsedPartStat.Total, "freeBytes": mostUsedPartStat.Free}).Infof("highest disk usage: %2.f%%", mostUsedPartStat.UsedPercent)
	}

	ioCounters, err := disk.IOCounters()
	if err != nil {
		return nil, fmt.Errorf("failed getting disk io counterss: %w", err)
	}
	for _, c := range ioCounters {
		_ = c
		systemData.DiskNodeIOSeconds = c.IoTime
		systemData.DiskNodeReadsTotal = c.ReadCount   // c.MergedReadCount ?
		systemData.DiskNodeWritesTotal = c.WriteCount // c.MergedWriteCount ?
	}

	netCounters, err := net.IOCounters(false)
	if err != nil {
		return nil, fmt.Errorf("failed getting net io counters: %w", err)
	}
	if len(netCounters) == 0 {
		return nil, fmt.Errorf("no net.IOCounters")
	}
	systemData.NetworkNodeBytesTotalReceive = netCounters[0].BytesRecv
	systemData.NetworkNodeBytesTotalTransmit = netCounters[0].BytesSent

	bootTime, err := host.BootTime()
	if err != nil {
		return nil, fmt.Errorf("failed getting boot time: %w", err)
	}
	systemData.MiscNodeBootTSSeconds = bootTime

	systemData.MiscOS = runtime.GOOS
	if len(systemData.MiscOS) > 3 {
		systemData.MiscOS = systemData.MiscOS[:3]
	}

	return systemData, nil
}

func getMetricValue(pb *promModel.Metric) float64 {
	if pb.Gauge != nil {
		return pb.Gauge.GetValue()
	}
	if pb.Counter != nil {
		return pb.Counter.GetValue()
	}
	if pb.Untyped != nil {
		return pb.Untyped.GetValue()
	}
	return math.NaN()
}

func getMetricValueFromFamilyMap(m map[string]*promModel.MetricFamily, name string) float64 {
	metricFamily, exists := m[name]
	if exists {
		m := metricFamily.GetMetric()
		if len(m) > 0 {
			return getMetricValue(m[0])
		}
	}
	return 0
}

func getMetrics(endpoint string) (map[string]*promModel.MetricFamily, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Cache-control", "no-cache")
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var parser promExpfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(res.Body)
	if err != nil {
		return nil, err
	}

	return metricFamilies, nil
}

func getPrysmBeaconnodeData(endpoint string, ts uint64) (*BeaconnodeData, error) {
	metrics, err := getMetrics(endpoint)
	if err != nil {
		return nil, err
	}

	data := &BeaconnodeData{}

	// CommonData
	data.Version = specVersion
	data.Timestamp = ts
	data.ExporterVersion = exporterVersion
	data.Process = "beaconnode"

	// ProcessData
	data.CPUProcessSecondsTotal = uint64(getMetricValueFromFamilyMap(metrics, "process_cpu_seconds_total"))
	data.MemoryProcessBytes = uint64(getMetricValueFromFamilyMap(metrics, "process_resident_memory_bytes"))
	data.ClientName = "prysm"

	prysmVersionMetric, exists := metrics["prysm_version"]
	if exists {
		ms := prysmVersionMetric.GetMetric()
		if len(ms) > 0 {
			ls := ms[0].GetLabel()
			for _, l := range ls {
				if l.Name != nil && l.Value != nil && *l.Name == "version" {
					data.ClientVersion = *l.Value
					break
				}
			}
		}
	}

	data.ClientBuild = 0
	data.SyncEth2FallbackConfigured = false
	data.SyncEth2FallbackConnected = false

	// BeaconnodeData
	data.DiskBeaconchainBytesTotal = uint64(getMetricValueFromFamilyMap(metrics, "bcnode_disk_beaconchain_bytes_total"))

	p2pMessageReceivedTotalMetric, exists := metrics["p2p_message_received_total"]
	if exists {
		ms := p2pMessageReceivedTotalMetric.GetMetric()
		total := uint64(0)
		for _, m := range ms {
			total += uint64(getMetricValue(m))
		}
		data.NetworkLibP2PBytesTotalReceive = total
	}

	data.NetworkLibP2PBytesTotalTransmit = 0

	p2pPeerCount, exists := metrics["p2p_peer_count"]
	if exists {
		ms := p2pPeerCount.GetMetric()
		for _, m := range ms {
			ls := m.GetLabel()
			for _, l := range ls {
				if l.Name != nil && l.Value != nil && *l.Name == "State" && *l.Value == "Connected" {
					data.ClientVersion = *l.Value
					data.NetworkPeersConnected = uint64(getMetricValue(m))
					break
				}
			}
		}
	}

	data.SyncEth1Connected = false
	data.SyncEth2Synced = false
	data.SyncBeaconHeadSlot = uint64(getMetricValueFromFamilyMap(metrics, "beacon_head_slot"))
	data.SyncEth1FallbackConfigured = false
	data.SyncEth1FallbackConnected = false
	data.SlasherActive = false

	return data, nil
}

func getPrysmValidatorData(endpoint string, ts uint64) (*ValidatorData, error) {
	metrics, err := getMetrics(endpoint)
	if err != nil {
		return nil, err
	}

	data := &ValidatorData{}

	// CommonData
	data.Version = specVersion
	data.Timestamp = ts
	data.ExporterVersion = exporterVersion
	data.Process = "validator"

	// ProcessData
	data.CPUProcessSecondsTotal = uint64(getMetricValueFromFamilyMap(metrics, "process_cpu_seconds_total"))
	data.MemoryProcessBytes = uint64(getMetricValueFromFamilyMap(metrics, "process_resident_memory_bytes"))
	data.ClientName = "prysm"

	prysmVersionMetric, exists := metrics["prysm_version"]
	if exists {
		ms := prysmVersionMetric.GetMetric()
		if len(ms) > 0 {
			ls := ms[0].GetLabel()
			for _, l := range ls {
				if l.Name != nil && l.Value != nil && *l.Name == "version" {
					data.ClientVersion = *l.Value
					break
				}
			}
		}
	}

	data.ClientBuild = 0
	data.SyncEth2FallbackConfigured = false
	data.SyncEth2FallbackConnected = false

	// ValidatorData

	validatorStatuses, exists := metrics["validator_statuses"]
	if exists {
		ms := validatorStatuses.GetMetric()
		for _, m := range ms {
			data.ValidatorTotal++
			if getMetricValue(m) == 3 {
				data.ValidatorActive++
			}
		}
	}

	return data, nil
}

func getNimbusBeaconnodeData(endpoint string, ts uint64) (*BeaconnodeData, error) {
	metrics, err := getMetrics(endpoint)
	if err != nil {
		return nil, err
	}

	data := &BeaconnodeData{}

	// CommonData
	data.Version = specVersion
	data.Timestamp = ts
	data.ExporterVersion = exporterVersion
	data.Process = "beaconnode"

	// ProcessData
	data.CPUProcessSecondsTotal = uint64(getMetricValueFromFamilyMap(metrics, "process_cpu_seconds_total"))
	data.MemoryProcessBytes = uint64(getMetricValueFromFamilyMap(metrics, "process_resident_memory_bytes"))
	data.ClientName = "nimbus"

	versionMetric, exists := metrics["version"]
	if exists {
		ms := versionMetric.GetMetric()
		if len(ms) > 0 {
			ls := ms[0].GetLabel()
			for _, l := range ls {
				if l.Name != nil && l.Value != nil && *l.Name == "version" {
					data.ClientVersion = *l.Value
					break
				}
			}
		}
	}

	// data.ClientBuild = 0
	// data.SyncEth2FallbackConfigured = false
	// data.SyncEth2FallbackConnected = false

	// // BeaconnodeData
	// data.DiskBeaconchainBytesTotal = uint64(getMetricValueFromFamilyMap(metrics, "bcnode_disk_beaconchain_bytes_total"))

	// p2pMessageReceivedTotalMetric, exists := metrics["p2p_message_received_total"]
	// if exists {
	// ms := p2pMessageReceivedTotalMetric.GetMetric()
	// total := uint64(0)
	// for _, m := range ms {
	// total += uint64(getMetricValue(m))
	// }
	// data.NetworkLibP2PBytesTotalReceive = total
	// }

	// data.NetworkLibP2PBytesTotalTransmit = 0

	// p2pPeerCount, exists := metrics["p2p_peer_count"]
	// if exists {
	// ms := p2pPeerCount.GetMetric()
	// for _, m := range ms {
	// ls := m.GetLabel()
	// for _, l := range ls {
	// if l.Name != nil && l.Value != nil && *l.Name == "State" && *l.Value == "Connected" {
	// data.ClientVersion = *l.Value
	// data.NetworkPeersConnected = uint64(getMetricValue(m))
	// break
	// }
	// }
	// }
	// }

	// data.SyncEth1Connected = false
	// data.SyncEth2Synced = false
	data.SyncBeaconHeadSlot = uint64(getMetricValueFromFamilyMap(metrics, "beacon_head_slot"))
	// data.SyncEth1FallbackConfigured = false
	// data.SyncEth1FallbackConnected = false
	// data.SlasherActive = false

	return data, nil
}
