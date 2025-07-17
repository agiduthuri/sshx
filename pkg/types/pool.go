// Package types - pool.go defines connection pool identification and management
// Connection pools group related connections and provide unique identification
package types

import (
	"fmt"
)

// PoolId represents a unique identifier for a connection pool
// Connection pools group related connections for the same application/service
// and provide tracking and management capabilities
type PoolId struct {
	// Value is the unique numeric identifier for this pool (typically timestamp)
	Value int64
	
	// Direction indicates the connection direction (inbound/outbound)
	Direction int32
	
	// ImplCode identifies which application implementation this pool serves
	// (APP_TYPE_SSH, APP_TYPE_VNC, etc.)
	ImplCode int32
}

// NewPoolId creates a new pool identifier with the given ID and implementation code
// id: Unique identifier (typically timestamp in nanoseconds)
// impc: Implementation code (APP_TYPE_* constant)
func NewPoolId(id int64, impc int32) *PoolId {
	return &PoolId{
		Value:    id,
		ImplCode: impc,
	}
}

// String returns a human-readable string representation of the pool ID
// Format: "conn_{ImplCode}_{Value}_{Direction}"
// This is used for logging and debugging purposes
func (pd *PoolId) String(direct int32) string {
	return fmt.Sprintf("conn_%d_%d_%d", pd.ImplCode, pd.Value, direct)
}

// Raw returns the raw numeric value of the pool ID
// This is used for comparisons and as a unique identifier
func (pd *PoolId) Raw() int64 {
	return pd.Value
}
