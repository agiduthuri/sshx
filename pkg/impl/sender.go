package impl

// Package impl provides the sender functionality for communicating with the sshx daemon.
// The Sender struct acts as the primary communication protocol between client applications
// and the TCP daemon running on LocalTCPPort (default: 2224).
//
// Communication Flow:
// 1. Client creates an application implementation (SSH, VNC, Proxy, etc.)
// 2. NewSender() encodes the implementation and creates a Sender request
// 3. Sender.Send() connects to the daemon via TCP and sends the gob-encoded request
// 4. Daemon (internal/node/tcp.go) receives, decodes, and processes the request
// 5. Connection is established via WebRTC or direct TCP based on the implementation
//
// Usage Examples:
// - SSH: Creates SSH connection via NewSender(sshImpl, OPTION_TYPE_UP)
// - Proxy: Creates proxy tunnel via NewSender(proxyImpl, OPTION_TYPE_UP) in doDial()
// - File Transfer: Manages file transfers via various OPTION_TYPE_UP/DOWN operations
// - SCP/SSHFS: Establishes file system connections

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"github.com/suutaku/sshx/pkg/conf"
)

// Sender represents a request structure sent to the Local TCP daemon (internal/node/tcp.go).
// This is the primary IPC mechanism between client applications and the sshx daemon.
// The daemon listens on LocalTCPPort (default: 2224) and processes these requests.
type Sender struct {
	// Type contains both the application code (upper bits) and option code (lower 8 bits)
	// Format: (appCode << flagLen) | optionCode
	// Application codes: APP_TYPE_SSH, APP_TYPE_VNC, APP_TYPE_PROXY, etc. (pkg/types/types.go)
	// Option codes: OPTION_TYPE_UP, OPTION_TYPE_DOWN, OPTION_TYPE_STAT, OPTION_TYPE_ATTACH
	Type       int32
	
	// PairId uniquely identifies a connection pair for matching client/server sides
	// Generated from the application's PairId() method, used for connection tracking
	PairId     []byte
	
	// Detach indicates whether the connection should be detached after establishment
	// When true, the sender doesn't wait for the connection to complete
	Detach     bool
	
	// LocalEntry is the TCP daemon address (typically "127.0.0.1:2224")
	// Automatically set from LocalTCPPort configuration
	LocalEntry string
	
	// Payload contains the gob-encoded application implementation (SSH, VNC, Proxy, etc.)
	// The daemon decodes this to get the specific application configuration
	Payload    []byte
	
	// Status is set by the daemon to indicate success (0) or failure (non-zero)
	// Only valid in response messages from the daemon
	Status     int32
}

// NewSender creates a new Sender request for the given application implementation and option code.
// This is the primary constructor used by all applications (SSH, VNC, Proxy, SCP, etc.)
//
// Parameters:
//   - imp: The application implementation (must implement the Impl interface)
//   - optCode: Operation type (OPTION_TYPE_UP, OPTION_TYPE_DOWN, etc.)
//
// Usage Examples:
//   - NewSender(sshImpl, types.OPTION_TYPE_UP) - Establish SSH connection
//   - NewSender(proxyImpl, types.OPTION_TYPE_UP) - Start proxy tunnel (used in proxy.doDial())
//   - NewSender(transferImpl, types.OPTION_TYPE_DOWN) - Close file transfer
//
// The function:
//   1. Encodes the app type and option code into a single Type field
//   2. Serializes the implementation using gob encoding
//   3. Sets up the daemon address from LocalTCPPort configuration
//   4. Copies the connection PairId for tracking
func NewSender(imp Impl, optCode int32) *Sender {
	// Warn if HostId is not set - this may cause connection issues
	if imp.HostId() == "" {
		logrus.Warn("Host Id not set, maybe you should set it on Preper")
	}
	
	// Combine application code (upper bits) with option code (lower 8 bits)
	// flagLen is defined in impl.go and determines the bit shift amount
	ret := &Sender{
		Type: (imp.Code() << flagLen) | optCode,
	}
	
	// Serialize the application implementation using gob encoding
	// This allows the daemon to reconstruct the exact application configuration
	buf := bytes.Buffer{}
	err := gob.NewEncoder(&buf).Encode(imp)
	if err != nil {
		logrus.Error(err)
		return nil
	}
	ret.Payload = buf.Bytes()
	
	// Set the daemon TCP address from configuration (default: 127.0.0.1:2224)
	cm := conf.NewConfManager("")
	ret.LocalEntry = fmt.Sprintf("127.0.0.1:%d", cm.Conf.LocalTCPPort)
	
	// Copy the connection pair ID for tracking this specific connection
	ret.PairId = []byte(imp.PairId())
	return ret
}

// GetAppCode extracts the application type code from the encoded Type field.
// Used by the daemon to determine which application type is being requested.
//
// Returns application codes like:
//   - APP_TYPE_SSH (0) for SSH connections
//   - APP_TYPE_VNC (1) for VNC connections  
//   - APP_TYPE_PROXY (4) for proxy tunnels
//   - APP_TYPE_TRANSFER (9) for file transfers
//
// The daemon uses this to route requests to the appropriate connection manager.
func (sender *Sender) GetAppCode() int32 {
	// Extract upper bits by right-shifting by flagLen
	return sender.Type >> flagLen
}

// GetOptionCode extracts the option type code from the encoded Type field.
// Used by the daemon (internal/node/tcp.go) to determine the operation type.
//
// Returns option codes like:
//   - OPTION_TYPE_UP (0) - Establish/bring up connection
//   - OPTION_TYPE_DOWN (1) - Tear down/close connection
//   - OPTION_TYPE_STAT (2) - Query connection status  
//   - OPTION_TYPE_ATTACH (3) - Attach to existing connection
//
// The daemon uses this to route requests to appropriate handlers in the ConnectionManager.
func (sender *Sender) GetOptionCode() int32 {
	// Extract lower 8 bits using bit mask
	return sender.Type & 0xff
}

// GetImpl deserializes and reconstructs the application implementation from the Payload.
// Used by the daemon to extract the original application configuration that was encoded
// in NewSender(). This allows the daemon to access all application-specific settings.
//
// Process:
//   1. Creates empty implementation based on application code
//   2. Deserializes the gob-encoded payload into the implementation
//   3. Returns the fully configured implementation
//
// The daemon uses this to get application details like:
//   - SSH: target host, port, authentication details
//   - Proxy: proxy port, target host ID
//   - VNC: VNC server details and connection parameters
func (sender *Sender) GetImpl() Impl {
	// Create empty implementation instance based on application type
	impl := GetImpl(sender.GetAppCode())
	
	// Deserialize the gob-encoded payload back into the implementation
	buf := bytes.NewBuffer(sender.Payload)
	err := gob.NewDecoder(buf).Decode(impl)
	if err != nil {
		logrus.Error(err)
	}
	return impl
}

// Send establishes a TCP connection to the daemon and sends this Sender request.
// This is the main communication method used by all applications to interact with the daemon.
//
// Communication Protocol:
//   1. Connect to daemon via TCP on LocalEntry address (127.0.0.1:2224)
//   2. Send gob-encoded Sender request to daemon
//   3. Wait for daemon to process request and send response
//   4. Receive updated Sender with Status field set
//   5. Return the TCP connection if successful
//
// Usage Examples:
//   - SSH: conn, err := NewSender(sshImpl, OPTION_TYPE_UP).Send()
//   - Proxy: conn, err := NewSender(proxyImpl, OPTION_TYPE_UP).Send() (in doDial)
//   - File Transfer: conn, err := NewSender(transferImpl, OPTION_TYPE_UP).Send()
//
// The returned connection can then be used for:
//   - Data piping (utils.Pipe for proxy connections)
//   - Direct application communication (SSH, file transfers)
//   - Connection management and cleanup
//
// Returns:
//   - net.Conn: Active TCP connection to daemon for data transfer
//   - error: Connection or protocol error
func (sender *Sender) Send() (net.Conn, error) {
	// Connect to the local daemon via TCP
	conn, err := net.Dial("tcp", sender.LocalEntry)
	if err != nil {
		return nil, err
	}
	
	// Send the gob-encoded request to daemon
	err = gob.NewEncoder(conn).Encode(sender)
	if err != nil {
		return nil, err
	}
	logrus.Debug("waiting TCP Responnse")

	// Wait for daemon response - daemon will update Status field
	err = gob.NewDecoder(conn).Decode(sender)
	if err != nil {
		return nil, err
	}

	logrus.Debug("TCP Responnse OK ", string(sender.PairId))
	
	// Check if daemon successfully processed the request
	if sender.Status != 0 {
		return nil, fmt.Errorf("response error")
	}
	
	// Return the active TCP connection for data transfer
	return conn, nil
}

// SendDetach sends the request with Detach flag set to true.
// Used for operations that don't need to maintain the connection after establishment.
//
// When Detach=true:
//   - The daemon processes the request but doesn't expect ongoing communication
//   - Useful for cleanup operations (OPTION_TYPE_DOWN) 
//   - Fire-and-forget style requests
//
// Usage Examples:
//   - File transfer cleanup: NewSender(impl, OPTION_TYPE_DOWN).SendDetach()
//   - Connection teardown: sender.SendDetach() when closing connections
//
// Currently used by:
//   - impl_transfer_service.go: For closing transfer services
//   - impl_sshfs.go: For cleanup operations (commented out)
//   - impl_proxy.go: For connection cleanup (commented out)
//
// Returns the same as Send() but with detached semantics.
func (sender *Sender) SendDetach() (net.Conn, error) {
	// Set the detach flag to indicate this is a detached operation
	sender.Detach = true
	// Use the standard Send() method with Detach=true
	return sender.Send()
}
