package main

// see: https://docs.google.com/document/d/1qPWAVRjPCENlyAjUBwGkHMvz9qLdd_6u9DPZcNxDBpc

type CommonData struct {
	Version         int64  `json:"version"`
	Timestamp       uint64 `json:"timestamp"` // unix timestamp in milliseconds
	Process         string `json:"process"`   // can be one of: validator, beaconnode, system
	ExporterVersion string `json:"exporter_version"`
}

type ProcessData struct {
	CommonData
	CPUProcessSecondsTotal     uint64 `json:"cpu_process_seconds_total"`
	MemoryProcessBytes         uint64 `json:"memory_process_bytes"`
	ClientName                 string `json:"client_name"` // can be one of: prysm, lighthouse, nimbus, teku
	ClientVersion              string `json:"client_version"`
	ClientBuild                int64  `json:"client_build"`
	SyncEth2FallbackConfigured bool   `json:"sync_eth2_fallback_configured"`
	SyncEth2FallbackConnected  bool   `json:"sync_eth2_fallback_connected"`
}

type BeaconnodeData struct {
	ProcessData
	DiskBeaconchainBytesTotal       uint64 `json:"disk_beaconchain_bytes_total"`
	NetworkLibP2PBytesTotalReceive  uint64 `json:"network_libp2p_bytes_total_receive"`
	NetworkLibP2PBytesTotalTransmit uint64 `json:"network_libp2p_bytes_total_transmit"`
	NetworkPeersConnected           uint64 `json:"network_peers_connected"`
	SyncEth1Connected               bool   `json:"sync_eth1_connected"`
	SyncEth2Synced                  bool   `json:"sync_eth2_synced"`
	SyncBeaconHeadSlot              uint64 `json:"sync_beacon_head_slot"`
	SyncEth1FallbackConfigured      bool   `json:"sync_eth1_fallback_configured"`
	SyncEth1FallbackConnected       bool   `json:"sync_eth1_fallback_connected"`
	SlasherActive                   bool   `json:"slasher_active"`
}

type ValidatorData struct {
	ProcessData
	ValidatorTotal  int64 `json:"validator_total"`
	ValidatorActive int64 `json:"validator_active"`
}

type SystemData struct {
	CommonData
	CPUCores                      int64  `json:"cpu_cores"`
	CPUThreads                    int64  `json:"cpu_threads"`
	CPUNodeSystemSecondsTotal     uint64 `json:"cpu_node_system_seconds_total"`
	CPUNodeUserSecondsTotal       uint64 `json:"cpu_node_user_seconds_total"`
	CPUNodeIOWaitSecondsTotal     uint64 `json:"cpu_node_iowait_seconds_total"`
	CPUNodeIdleSecondsTotal       uint64 `json:"cpu_node_idle_seconds_total"`
	MemoryNodeBytesTotal          uint64 `json:"memory_node_bytes_total"`
	MemoryNodeBytesFree           uint64 `json:"memory_node_bytes_free"`
	MemoryNodeBytesCached         uint64 `json:"memory_node_bytes_cached"`
	MemoryNodeBytesBuffers        uint64 `json:"memory_node_bytes_buffers"`
	DiskNodeBytesTotal            uint64 `json:"disk_node_bytes_total"`
	DiskNodeBytesFree             uint64 `json:"disk_node_bytes_free"`
	DiskNodeIOSeconds             uint64 `json:"disk_node_io_seconds"`
	DiskNodeReadsTotal            uint64 `json:"disk_node_reads_total"`
	DiskNodeWritesTotal           uint64 `json:"disk_node_writes_total"`
	NetworkNodeBytesTotalReceive  uint64 `json:"network_node_bytes_total_receive"`
	NetworkNodeBytesTotalTransmit uint64 `json:"network_node_bytes_total_transmit"`
	MiscNodeBootTSSeconds         uint64 `json:"misc_node_boot_ts_seconds"`
	MiscOS                        string `json:"misc_os"`
}
