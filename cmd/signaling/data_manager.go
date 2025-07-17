package main

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/suutaku/sshx/pkg/types"
)

const (
	LIFE_TIME_IN_SECOND = 15 // Timeout in seconds before cleaning up inactive peers
	MAX_BUFFER_NUMBER   = 64 // Maximum number of queued messages per peer
)

// DManager (Data Manager) handles peer message queues and lifecycle management
// It maintains a map of peer IDs to their message channels and implements
// automatic cleanup of inactive peers using a watchdog mechanism
type DManager struct {
	datas map[string]chan types.SignalingInfo // Message channels for each peer ID
	mu    sync.Mutex                          // Mutex for thread-safe access to maps
	alive map[string]int                      // Keepalive counters for each peer (in seconds)
}

// NewDManager creates a new data manager instance
// Initializes empty maps for peer data channels and keepalive counters
func NewDManager() *DManager {
	return &DManager{
		datas: make(map[string]chan types.SignalingInfo),
		alive: make(map[string]int),
	}
}

// Get retrieves the message channel for a specific peer ID
// Returns nil if the peer doesn't exist (no messages queued)
// This is used by the pull endpoint to check for available messages
func (dm *DManager) Get(id string) chan types.SignalingInfo {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	return dm.datas[id]
}

// Clean removes a peer from the data manager
// Closes the peer's message channel and removes them from both maps
// This prevents memory leaks from inactive peers
func (dm *DManager) Clean(id string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	// Close the channel if it exists to prevent goroutine leaks
	if dm.datas[id] != nil {
		close(dm.datas[id])
	}
	
	// Remove peer from both tracking maps
	delete(dm.datas, id)
	delete(dm.alive, id)
}

// resetAlive resets the keepalive counter for a peer to maximum lifetime
// Called whenever a peer shows activity (sends/receives messages)
func (dm *DManager) resetAlive(id string) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.alive[id] = LIFE_TIME_IN_SECOND
}

// Set queues a message for a specific peer and manages their lifecycle
// Creates a new peer entry if they don't exist, including starting a watchdog
// Uses a buffered channel to prevent blocking when multiple messages arrive
func (dm *DManager) Set(id string, info types.SignalingInfo) {
	// Create new peer entry if it doesn't exist
	if dm.datas[id] == nil {
		dm.mu.Lock()
		// Create buffered channel to queue messages for this peer
		dm.datas[id] = make(chan types.SignalingInfo, MAX_BUFFER_NUMBER)
		dm.mu.Unlock()
		
		// Initialize keepalive timer
		dm.resetAlive(id)
		
		// Start watchdog goroutine for automatic cleanup
		go func(dmc *DManager) {
			logrus.Debug("create watch dog for ", id)
			
			// Countdown timer - decrements every second
			for dmc.alive[id] > 0 {
				time.Sleep(time.Second)
				dmc.mu.Lock()
				dmc.alive[id]--
				dmc.mu.Unlock()
			}
			
			// Timer expired - clean up this peer
			logrus.Debug("execute watch dog for ", id)
			dm.Clean(id)
		}(dm)
	}
	
	// Try to queue the message (non-blocking)
	select {
	case dm.datas[id] <- info:
		// Message queued successfully - reset keepalive timer
		dm.resetAlive(id)
	default:
		// Channel full - message dropped
		// This prevents the server from blocking on slow peers
	}
}
