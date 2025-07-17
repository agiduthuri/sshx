// Package node provides the core node functionality for the sshx system
// The Node is the central component that manages configuration, connections, and services
package node

import (
	"github.com/suutaku/sshx/internal/conn"
	"github.com/suutaku/sshx/pkg/conf"
)

// Node represents the main sshx node that coordinates all system components
// It manages configuration, connection services, and provides the TCP server interface
type Node struct {
	// confManager handles configuration loading, saving, and live reloading
	confManager *conf.ConfManager
	
	// running indicates whether the node is currently active
	running bool
	
	// connMgr manages all connection services (direct TCP and WebRTC)
	connMgr *conn.ConnectionManager
}

func NewNode(home string) *Node {
	cm := conf.NewConfManager(home)
	enabledService := []conn.ConnectionService{
		conn.NewDirectService(cm.Conf.ID),
		conn.NewWebRTCService(cm.Conf.ID, cm.Conf.SignalingServerAddr, cm.Conf.RTCConf),
	}
	return &Node{
		confManager: cm,
		connMgr:     conn.NewConnectionManager(enabledService),
	}
}

func (node *Node) Start() {
	node.running = true
	go node.connMgr.Start()
	node.ServeTCP()
}

func (node *Node) Stop() {
	node.running = false
	node.connMgr.Stop()
}
