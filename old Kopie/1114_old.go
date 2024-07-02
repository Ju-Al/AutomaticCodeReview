// Copyright (C) 2019-2020 Algorand, Inc.
// This file is part of go-algorand
//
// go-algorand is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// go-algorand is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with go-algorand.  If not, see <https://www.gnu.org/licenses/>.

package config

import (
	"fmt"
var defaultLocal = defaultLocalV7

const configVersion = uint32(7)

// !!! WARNING !!!
//
// These versioned structures need to be maintained CAREFULLY and treated
// like UNIVERSAL CONSTANTS - they should not be modified once committed.
//
// New fields may be added to the current defaultLocalV# and should
// also be added to installer/config.json.example and
// test/testdata/configs/config-v{n}.json
//
// Changing a default value requires creating a new defaultLocalV# instance,
// bump the version number (configVersion), and add appropriate migration and tests.
//
// !!! WARNING !!!

var defaultLocalV7 = Local{
	// DO NOT MODIFY VALUES - New values may be added carefully - See WARNING at top of file
	Version:                               7,
	Archival:                              false,
	BaseLoggerDebugLevel:                  4,
	BroadcastConnectionsLimit:             -1,
	AnnounceParticipationKey:              true,
	PriorityPeers:                         map[string]bool{},
	CadaverSizeTarget:                     1073741824,
	CatchupFailurePeerRefreshRate:         10,
	CatchupParallelBlocks:                 16,
	ConnectionsRateLimitingCount:          60,
	ConnectionsRateLimitingWindowSeconds:  1,
	DeadlockDetection:                     0,
	DNSBootstrapID:                        "<network>.algorand.network",
	EnableAgreementReporting:              false,
	EnableAgreementTimeMetrics:            false,
	EnableIncomingMessageFilter:           false,
	EnableMetricReporting:                 false,
	EnableOutgoingNetworkMessageFiltering: true,
	EnableRequestLogger:                   false,
	EnableTopAccountsReporting:            false,
	EndpointAddress:                       "127.0.0.1:0",
	GossipFanout:                          4,
	IncomingConnectionsLimit:              10000,
	IncomingMessageFilterBucketCount:      5,
	IncomingMessageFilterBucketSize:       512,
	LogArchiveName:                        "node.archive.log",
	LogArchiveMaxAge:                      "",
	LogSizeLimit:                          1073741824,
	MaxConnectionsPerIP:                   30,
	NetAddress:                            "",
	NodeExporterListenAddress:             ":9100",
	NodeExporterPath:                      "./node_exporter",
	OutgoingMessageFilterBucketCount:      3,
	OutgoingMessageFilterBucketSize:       128,
	ReconnectTime:                         1 * time.Minute,
	ReservedFDs:                           256,
	RestReadTimeoutSeconds:                15,
	RestWriteTimeoutSeconds:               120,
	RunHosted:                             false,
	SuggestedFeeBlockHistory:              3,
	SuggestedFeeSlidingWindowSize:         50,
	TelemetryToLog:                        true,
	TxPoolExponentialIncreaseFactor:       2,
	TxPoolSize:                            15000,
	TxSyncIntervalSeconds:                 60,
	TxSyncTimeoutSeconds:                  30,
	TxSyncServeResponseSize:               1000000,
	PeerConnectionsUpdateInterval:         3600,
	DNSSecurityFlags:                      0x01,
	EnablePingHandler:                     true,
	CatchpointInterval:                    10000, // added in V7
	CatchpointFileHistoryLength:           365,   // add in V7
	EnableLedgerService:                   false, // added in V7
	EnableBlockService:                    false, // added in V7
	// DO NOT MODIFY VALUES - New values may be added carefully - See WARNING at top of file
}

var defaultLocalV6 = Local{
	// DO NOT MODIFY VALUES - New values may be added carefully - See WARNING at top of file
	Version:                               6,
	Archival:                              false,
	BaseLoggerDebugLevel:                  4,
	BroadcastConnectionsLimit:             -1,
	AnnounceParticipationKey:              true,
	PriorityPeers:                         map[string]bool{},
	CadaverSizeTarget:                     1073741824,
	CatchupFailurePeerRefreshRate:         10,
	CatchupParallelBlocks:                 16,
	ConnectionsRateLimitingCount:          60,
	ConnectionsRateLimitingWindowSeconds:  1,
	DeadlockDetection:                     0,
	DNSBootstrapID:                        "<network>.algorand.network",
	EnableAgreementReporting:              false,
	EnableAgreementTimeMetrics:            false,
	EnableIncomingMessageFilter:           false,
	EnableMetricReporting:                 false,
	EnableOutgoingNetworkMessageFiltering: true,
	EnableRequestLogger:                   false,
	EnableTopAccountsReporting:            false,
	EndpointAddress:                       "127.0.0.1:0",
	GossipFanout:                          4,
	IncomingConnectionsLimit:              10000,
	IncomingMessageFilterBucketCount:      5,
	IncomingMessageFilterBucketSize:       512,
	LogArchiveName:                        "node.archive.log",
	LogArchiveMaxAge:                      "",
	LogSizeLimit:                          1073741824,
	MaxConnectionsPerIP:                   30,
	NetAddress:                            "",
	NetworkProtocolVersion:                "",
	NodeExporterListenAddress:             ":9100",
	NodeExporterPath:                      "./node_exporter",
	OutgoingMessageFilterBucketCount:      3,
	OutgoingMessageFilterBucketSize:       128,
	ReconnectTime:                         1 * time.Minute,
	ReservedFDs:                           256,
	RestReadTimeoutSeconds:                15,
	RestWriteTimeoutSeconds:               120,
	RunHosted:                             false,
	SuggestedFeeBlockHistory:              3,
	SuggestedFeeSlidingWindowSize:         50,
	TelemetryToLog:                        true,
	TxPoolExponentialIncreaseFactor:       2,
	TxPoolSize:                            15000,
	TxSyncIntervalSeconds:                 60,
	TxSyncTimeoutSeconds:                  30,
	TxSyncServeResponseSize:               1000000,
	PeerConnectionsUpdateInterval:         3600,
	DNSSecurityFlags:                      0x01, // New value with default 0x01
	EnablePingHandler:                     true,
	// DO NOT MODIFY VALUES - New values may be added carefully - See WARNING at top of file
}

var defaultLocalV5 = Local{
	// DO NOT MODIFY VALUES - New values may be added carefully - See WARNING at top of file
	Version:                               5,
	Archival:                              false,
	BaseLoggerDebugLevel:                  4, // Was 1
	BroadcastConnectionsLimit:             -1,
	AnnounceParticipationKey:              true,
	PriorityPeers:                         map[string]bool{},
	CadaverSizeTarget:                     1073741824,
	CatchupFailurePeerRefreshRate:         10,
	CatchupParallelBlocks:                 16,
	ConnectionsRateLimitingCount:          60,
	ConnectionsRateLimitingWindowSeconds:  1,
	DeadlockDetection:                     0,
	DisableOutgoingConnectionThrottling:   false,
	DNSBootstrapID:                        "<network>.algorand.network",
	EnableAgreementReporting:              false,
	EnableAgreementTimeMetrics:            false,
	EnableIncomingMessageFilter:           false,
	EnableMetricReporting:                 false,
	EnableOutgoingNetworkMessageFiltering: true,
	EnableRequestLogger:                   false,
	EnableTopAccountsReporting:            false,
	EndpointAddress:                       "127.0.0.1:0",
	GossipFanout:                          4,
	IncomingConnectionsLimit:              10000, // Was -1
	IncomingMessageFilterBucketCount:      5,
	IncomingMessageFilterBucketSize:       512,
	LogArchiveName:                        "node.archive.log",
	LogArchiveMaxAge:                      "",
	LogSizeLimit:                          1073741824,
	MaxConnectionsPerIP:                   30,
	NetAddress:                            "",
	NodeExporterListenAddress:             ":9100",
	NodeExporterPath:                      "./node_exporter",
	OutgoingMessageFilterBucketCount:      3,
	OutgoingMessageFilterBucketSize:       128,
	PeerConnectionsUpdateInterval:         3600,
	ReconnectTime:                         1 * time.Minute, // Was 60ns
	ReservedFDs:                           256,
	RestReadTimeoutSeconds:                15,
	RestWriteTimeoutSeconds:               120,
	RunHosted:                             false,
	SuggestedFeeBlockHistory:              3,
	SuggestedFeeSlidingWindowSize:         50,
	TelemetryToLog:                        true,
	TxPoolExponentialIncreaseFactor:       2,
	TxPoolSize:                            15000,
	TxSyncIntervalSeconds:                 60,
	TxSyncTimeoutSeconds:                  30,
	TxSyncServeResponseSize:               1000000,
	// DO NOT MODIFY VALUES - New values may be added carefully - See WARNING at top of file
}

var defaultLocalV4 = Local{
	// DO NOT MODIFY VALUES - New values may be added carefully - See WARNING at top of file
	Version:                               4,
	Archival:                              false,
	BaseLoggerDebugLevel:                  4, // Was 1
	BroadcastConnectionsLimit:             -1,
	AnnounceParticipationKey:              true,
	PriorityPeers:                         map[string]bool{},
	CadaverSizeTarget:                     1073741824,
	CatchupFailurePeerRefreshRate:         10,
	CatchupParallelBlocks:                 50,
	ConnectionsRateLimitingCount:          60,
	ConnectionsRateLimitingWindowSeconds:  1,
	DeadlockDetection:                     0,
	DNSBootstrapID:                        "<network>.algorand.network",
	EnableAgreementReporting:              false,
	EnableAgreementTimeMetrics:            false,
	EnableIncomingMessageFilter:           false,
	EnableMetricReporting:                 false,
	EnableOutgoingNetworkMessageFiltering: true,
	EnableRequestLogger:                   false,
	EnableTopAccountsReporting:            false,
	EndpointAddress:                       "127.0.0.1:0",
	GossipFanout:                          4,
	IncomingConnectionsLimit:              10000, // Was -1
	IncomingMessageFilterBucketCount:      5,
	IncomingMessageFilterBucketSize:       512,
	LogArchiveName:                        "node.archive.log",
	LogArchiveMaxAge:                      "",
	LogSizeLimit:                          1073741824,
	MaxConnectionsPerIP:                   30,
	NetAddress:                            "",
	NodeExporterListenAddress:             ":9100",
	NodeExporterPath:                      "./node_exporter",
	OutgoingMessageFilterBucketCount:      3,
	OutgoingMessageFilterBucketSize:       128,
	ReconnectTime:                         1 * time.Minute, // Was 60ns
	ReservedFDs:                           256,
	RestReadTimeoutSeconds:                15,
	RestWriteTimeoutSeconds:               120,
	RunHosted:                             false,
	SuggestedFeeBlockHistory:              3,
	SuggestedFeeSlidingWindowSize:         50,
	TxPoolExponentialIncreaseFactor:       2,
	TxPoolSize:                            50000,
	TxSyncIntervalSeconds:                 60,
	TxSyncTimeoutSeconds:                  30,
	TxSyncServeResponseSize:               1000000,

	// DO NOT MODIFY VALUES - New values may be added carefully - See WARNING at top of file
}

var defaultLocalV3 = Local{
	// DO NOT MODIFY VALUES - New values may be added carefully - See WARNING at top of file
	Version:                               3,
	Archival:                              false,
	BaseLoggerDebugLevel:                  4, // Was 1
	CadaverSizeTarget:                     1073741824,
	CatchupFailurePeerRefreshRate:         10,
	CatchupParallelBlocks:                 50,
	DeadlockDetection:                     0,
	DNSBootstrapID:                        "<network>.algorand.network",
	EnableAgreementReporting:              false,
	EnableAgreementTimeMetrics:            false,
	EnableIncomingMessageFilter:           false,
	EnableMetricReporting:                 false,
	EnableOutgoingNetworkMessageFiltering: true,
	EnableTopAccountsReporting:            false,
	EndpointAddress:                       "127.0.0.1:0",
	GossipFanout:                          4,
	IncomingConnectionsLimit:              10000, // Was -1
	IncomingMessageFilterBucketCount:      5,
	IncomingMessageFilterBucketSize:       512,
	LogSizeLimit:                          1073741824,
	MaxConnectionsPerIP:                   30,
	NetAddress:                            "",
	NodeExporterListenAddress:             ":9100",
	NodeExporterPath:                      "./node_exporter",
	OutgoingMessageFilterBucketCount:      3,
	OutgoingMessageFilterBucketSize:       128,
	ReconnectTime:                         1 * time.Minute, // Was 60ns
	ReservedFDs:                           256,
	RunHosted:                             false,
	SuggestedFeeBlockHistory:              3,
	SuggestedFeeSlidingWindowSize:         50,
	TxPoolExponentialIncreaseFactor:       2,
	TxPoolSize:                            50000,
	TxSyncIntervalSeconds:                 60,
	TxSyncTimeoutSeconds:                  30,
	TxSyncServeResponseSize:               1000000,
	IsIndexerActive:                       false,
	// DO NOT MODIFY VALUES - New values may be added carefully - See WARNING at top of file
}

var defaultLocalV2 = Local{
	// DO NOT MODIFY VALUES - New values may be added carefully - See WARNING at top of file
	Version:                               2,
	Archival:                              false,
	BaseLoggerDebugLevel:                  4, // Was 1
	CadaverSizeTarget:                     1073741824,
	CatchupFailurePeerRefreshRate:         10,
	DeadlockDetection:                     0,
	DNSBootstrapID:                        "<network>.algorand.network",
	EnableIncomingMessageFilter:           false,
	EnableMetricReporting:                 false,
	EnableOutgoingNetworkMessageFiltering: true,
	EnableTopAccountsReporting:            false,
	EndpointAddress:                       "127.0.0.1:0",
	GossipFanout:                          4,
	IncomingConnectionsLimit:              10000, // Was -1
	IncomingMessageFilterBucketCount:      5,
	IncomingMessageFilterBucketSize:       512,
	LogSizeLimit:                          1073741824,
	NetAddress:                            "",
	NodeExporterListenAddress:             ":9100",
	NodeExporterPath:                      "./node_exporter",
	OutgoingMessageFilterBucketCount:      3,
	OutgoingMessageFilterBucketSize:       128,
	ReconnectTime:                         1 * time.Minute, // Was 60ns
	ReservedFDs:                           256,
	SuggestedFeeBlockHistory:              3,
	TxPoolExponentialIncreaseFactor:       2,
	TxPoolSize:                            50000,
	TxSyncIntervalSeconds:                 60,
	TxSyncTimeoutSeconds:                  30,
	// DO NOT MODIFY VALUES - New values may be added carefully - See WARNING at top of file
}

var defaultLocalV1 = Local{
	// DO NOT MODIFY VALUES - New values may be added carefully - See WARNING at top of file
	Version:                               1,
	Archival:                              false,
	BaseLoggerDebugLevel:                  4, // Was 1
	CadaverSizeTarget:                     1073741824,
	CatchupFailurePeerRefreshRate:         10,
	DeadlockDetection:                     0,
	DNSBootstrapID:                        "<network>.algorand.network",
	EnableIncomingMessageFilter:           false,
	EnableMetricReporting:                 false,
	EnableOutgoingNetworkMessageFiltering: true,
	EnableTopAccountsReporting:            false,
	EndpointAddress:                       "127.0.0.1:0",
	GossipFanout:                          4,
	IncomingConnectionsLimit:              10000, // Was -1
	IncomingMessageFilterBucketCount:      5,
	IncomingMessageFilterBucketSize:       512,
	LogSizeLimit:                          1073741824,
	NetAddress:                            "",
	NodeExporterListenAddress:             ":9100",
	NodeExporterPath:                      "./node_exporter",
	OutgoingMessageFilterBucketCount:      3,
	OutgoingMessageFilterBucketSize:       128,
	ReconnectTime:                         1 * time.Minute, // Was 60ns
	SuggestedFeeBlockHistory:              3,
	TxPoolExponentialIncreaseFactor:       2,
	TxPoolSize:                            50000,
	TxSyncIntervalSeconds:                 60,
	TxSyncTimeoutSeconds:                  30,
	// DO NOT MODIFY VALUES - New values may be added carefully - See WARNING at top of file
}

var defaultLocalV0 = Local{
	// DO NOT MODIFY VALUES - New values may be added carefully - See WARNING at top of file
	Version:                               0,
	Archival:                              false,
	BaseLoggerDebugLevel:                  1,
	CadaverSizeTarget:                     1073741824,
	CatchupFailurePeerRefreshRate:         10,
	DNSBootstrapID:                        "<network>.algorand.network",
	EnableIncomingMessageFilter:           false,
	EnableMetricReporting:                 false,
	EnableOutgoingNetworkMessageFiltering: true,
	EnableTopAccountsReporting:            false,
	EndpointAddress:                       "127.0.0.1:0",
	GossipFanout:                          4,
	IncomingConnectionsLimit:              -1,
	IncomingMessageFilterBucketCount:      5,
	IncomingMessageFilterBucketSize:       512,
	LogSizeLimit:                          1073741824,
	NetAddress:                            "",
	NodeExporterListenAddress:             ":9100",
	NodeExporterPath:                      "./node_exporter",
	OutgoingMessageFilterBucketCount:      3,
	OutgoingMessageFilterBucketSize:       128,
	ReconnectTime:                         60,
	SuggestedFeeBlockHistory:              3,
	TxPoolExponentialIncreaseFactor:       2,
	TxPoolSize:                            50000,
	TxSyncIntervalSeconds:                 60,
	TxSyncTimeoutSeconds:                  30,
	// DO NOT MODIFY VALUES - New values may be added carefully - See WARNING at top of file
}
	if cfg.Version == configVersion {
		return
	}
	if cfg.Version > configVersion {
	// For now, manually perform migration.
	// When we have more time, we can use reflection to migrate from initial
	// version to latest version (progressively applying defaults)
	// Migrate 0 -> 1
	if newCfg.Version == 0 {
		if newCfg.BaseLoggerDebugLevel == defaultLocalV0.BaseLoggerDebugLevel {
			newCfg.BaseLoggerDebugLevel = defaultLocalV1.BaseLoggerDebugLevel
		if newCfg.IncomingConnectionsLimit == defaultLocalV0.IncomingConnectionsLimit {
			newCfg.IncomingConnectionsLimit = defaultLocalV1.IncomingConnectionsLimit
		}
		if newCfg.ReconnectTime == defaultLocalV0.ReconnectTime {
			newCfg.ReconnectTime = defaultLocalV1.ReconnectTime
		}
		newCfg.Version = 1
	}
	// Migrate 1 -> 2
	if newCfg.Version == 1 {
		if newCfg.ReservedFDs == defaultLocalV1.ReservedFDs {
			newCfg.ReservedFDs = defaultLocalV2.ReservedFDs
		}
		newCfg.Version = 2
	// Migrate 2 -> 3
	if newCfg.Version == 2 {
		if newCfg.MaxConnectionsPerIP == defaultLocalV2.MaxConnectionsPerIP {
			newCfg.MaxConnectionsPerIP = defaultLocalV3.MaxConnectionsPerIP
		}
		if newCfg.CatchupParallelBlocks == defaultLocalV2.CatchupParallelBlocks {
			newCfg.CatchupParallelBlocks = defaultLocalV3.CatchupParallelBlocks
		}
		newCfg.Version = 3
	"time"
	"strconv"
)

var defaultLocal = getVersionedDefaultLocalConfig(getLatestConfigVersion())

func migrate(cfg Local) (newCfg Local, err error) {
	newCfg = cfg
	latestConfigVersion := getLatestConfigVersion()

	if cfg.Version > latestConfigVersion {
		err = fmt.Errorf("unexpected config version: %d", cfg.Version)
		return
	}

	for {
		if newCfg.Version == latestConfigVersion {
			break
		}
		defaultCurrentConfig := getVersionedDefaultLocalConfig(newCfg.Version)
		localType := reflect.TypeOf(Local{})
		nextVersion := newCfg.Version + 1
		for fieldNum := 0; fieldNum < localType.NumField(); fieldNum++ {
			field := localType.Field(fieldNum)
			nextVersionDefaultValue, hasTag := reflect.StructTag(field.Tag).Lookup(fmt.Sprintf("version[%d]", nextVersion))
			if !hasTag {
				continue
			}
			if nextVersionDefaultValue == "" {
				switch reflect.ValueOf(&defaultCurrentConfig).Elem().FieldByName(field.Name).Kind() {
				case reflect.Map:
					// if the current implementation have a nil value, use the same value as
					// the default one ( i.e. empty map rather than nil map)
					if reflect.ValueOf(&newCfg).Elem().FieldByName(field.Name).Len() == 0 {
						reflect.ValueOf(&newCfg).Elem().FieldByName(field.Name).Set(reflect.MakeMap(field.Type))
					}
				case reflect.Array:
					// if the current implementation have a nil value, use the same value as
					// the default one ( i.e. empty slice rather than nil slice)
					if reflect.ValueOf(&newCfg).Elem().FieldByName(field.Name).Len() == 0 {
						reflect.ValueOf(&newCfg).Elem().FieldByName(field.Name).Set(reflect.MakeSlice(field.Type, 0, 0))
					}
				default:
				}
				continue
			}
			// we have found a field that has a new value for this new version. See if the current configuration value for that
			// field is identical to the default configuration for the field.
			switch reflect.ValueOf(&defaultCurrentConfig).Elem().FieldByName(field.Name).Kind() {
			case reflect.Bool:
				if reflect.ValueOf(&newCfg).Elem().FieldByName(field.Name).Bool() == reflect.ValueOf(&defaultCurrentConfig).Elem().FieldByName(field.Name).Bool() {
					// we're skipping the error checking here since we already tested that in the unit test.
					boolVal, _ := strconv.ParseBool(nextVersionDefaultValue)
					reflect.ValueOf(&newCfg).Elem().FieldByName(field.Name).SetBool(boolVal)
				}
			case reflect.Int32:
				fallthrough
			case reflect.Int:
				fallthrough
			case reflect.Int64:
				if reflect.ValueOf(&newCfg).Elem().FieldByName(field.Name).Int() == reflect.ValueOf(&defaultCurrentConfig).Elem().FieldByName(field.Name).Int() {
					// we're skipping the error checking here since we already tested that in the unit test.
					intVal, _ := strconv.ParseInt(nextVersionDefaultValue, 10, 64)
					reflect.ValueOf(&newCfg).Elem().FieldByName(field.Name).SetInt(intVal)
				}
			case reflect.Uint32:
				fallthrough
			case reflect.Uint:
				fallthrough
			case reflect.Uint64:
				if reflect.ValueOf(&newCfg).Elem().FieldByName(field.Name).Uint() == reflect.ValueOf(&defaultCurrentConfig).Elem().FieldByName(field.Name).Uint() {
					// we're skipping the error checking here since we already tested that in the unit test.
					uintVal, _ := strconv.ParseUint(nextVersionDefaultValue, 10, 64)
					reflect.ValueOf(&newCfg).Elem().FieldByName(field.Name).SetUint(uintVal)
				}
			case reflect.String:
				if reflect.ValueOf(&newCfg).Elem().FieldByName(field.Name).String() == reflect.ValueOf(&defaultCurrentConfig).Elem().FieldByName(field.Name).String() {
					// we're skipping the error checking here since we already tested that in the unit test.
					reflect.ValueOf(&newCfg).Elem().FieldByName(field.Name).SetString(nextVersionDefaultValue)
				}
			default:
				panic(fmt.Sprintf("unsupported data type (%s) encountered when reflecting on config.Local datatype %s", reflect.ValueOf(&defaultCurrentConfig).Elem().FieldByName(field.Name).Kind(), field.Name))
			}
		}
	}
	return
}

func getLatestConfigVersion() uint32 {
	localType := reflect.TypeOf(Local{})
	versionField, found := localType.FieldByName("Version")
	if !found {
		return 0
	}
	version := uint32(0)
	for {
		_, hasTag := reflect.StructTag(versionField.Tag).Lookup(fmt.Sprintf("version[%d]", version+1))
		if !hasTag {
			return version
		}
		version++
	}
}

func getVersionedDefaultLocalConfig(version uint32) (local Local) {
	if version < 0 {
		return
	}
	if version > 0 {
		local = getVersionedDefaultLocalConfig(version - 1)
	}
	// apply version specific changes.
	localType := reflect.TypeOf(local)
	for fieldNum := 0; fieldNum < localType.NumField(); fieldNum++ {
		field := localType.Field(fieldNum)
		versionDefaultValue, hasTag := reflect.StructTag(field.Tag).Lookup(fmt.Sprintf("version[%d]", version))
		if !hasTag {
			continue
		}
		if versionDefaultValue == "" {
			// set the default field value in case it's a map/array so we won't have nil ones.
			switch reflect.ValueOf(&local).Elem().FieldByName(field.Name).Kind() {
			case reflect.Map:
				reflect.ValueOf(&local).Elem().FieldByName(field.Name).Set(reflect.MakeMap(field.Type))
			case reflect.Array:
				reflect.ValueOf(&local).Elem().FieldByName(field.Name).Set(reflect.MakeSlice(field.Type, 0, 0))
			default:
			}
			continue
		}
		switch reflect.ValueOf(&local).Elem().FieldByName(field.Name).Kind() {
		case reflect.Bool:
			boolVal, err := strconv.ParseBool(versionDefaultValue)
			if err != nil {
				panic(err)
			}
			reflect.ValueOf(&local).Elem().FieldByName(field.Name).SetBool(boolVal)

		case reflect.Int32:
			intVal, err := strconv.ParseInt(versionDefaultValue, 10, 32)
			if err != nil {
				panic(err)
			}
			reflect.ValueOf(&local).Elem().FieldByName(field.Name).SetInt(intVal)
		case reflect.Int:
			fallthrough
		case reflect.Int64:
			intVal, err := strconv.ParseInt(versionDefaultValue, 10, 64)
			if err != nil {
				panic(err)
			}
			reflect.ValueOf(&local).Elem().FieldByName(field.Name).SetInt(intVal)

		case reflect.Uint32:
			uintVal, err := strconv.ParseUint(versionDefaultValue, 10, 32)
			if err != nil {
				panic(err)
			}
			reflect.ValueOf(&local).Elem().FieldByName(field.Name).SetUint(uintVal)
		case reflect.Uint:
			fallthrough
		case reflect.Uint64:
			uintVal, err := strconv.ParseUint(versionDefaultValue, 10, 64)
			if err != nil {
				panic(err)
			}
			reflect.ValueOf(&local).Elem().FieldByName(field.Name).SetUint(uintVal)
		case reflect.String:
			reflect.ValueOf(&local).Elem().FieldByName(field.Name).SetString(versionDefaultValue)
		default:
			panic(fmt.Sprintf("unsupported data type (%s) encountered when reflecting on config.Local datatype %s", reflect.ValueOf(&local).Elem().FieldByName(field.Name).Kind(), field.Name))
		}
	}
	return
}
