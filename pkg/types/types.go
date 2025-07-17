// Package types defines core type constants and structures used throughout the sshx system
// This includes application types, option types, and signaling types for WebRTC communication
package types

// Option types define the direction and purpose of connection operations
// These are used to indicate whether a connection is being established, torn down, or queried
const (
	OPTION_TYPE_UP     = iota // Establish/bring up a connection
	OPTION_TYPE_DOWN          // Tear down/close a connection
	OPTION_TYPE_STAT          // Query connection status
	OPTION_TYPE_ATTACH        // Attach to an existing connection
)

// Application types define the different services/applications supported by sshx
// Each application type corresponds to a specific implementation in pkg/impl/
const (
	APP_TYPE_SSH              = iota // SSH remote shell access
	APP_TYPE_VNC                     // VNC remote desktop client
	APP_TYPE_SCP                     // Secure copy file transfer
	APP_TYPE_SFS                     // SSH filesystem (SSHFS) mounting
	APP_TYPE_PROXY                   // HTTP/SOCKS proxy server
	APP_TYPE_STAT                    // Connection status monitoring
	APP_TYPE_VNC_SERVICE             // VNC server (provides VNC access)
	APP_TYPE_MESSAGER                // Real-time messaging console
	APP_TYPE_TRANSFER_SERVICE        // File transfer server
	APP_TYPE_TRANSFER                // File transfer client
)

// WebRTC signaling message types used in the peer-to-peer connection establishment
// These correspond to different phases of the WebRTC handshake process
const (
	SIG_TYPE_UNKNOWN   = iota // Unknown or invalid signaling message
	SIG_TYPE_CANDIDATE        // ICE candidate exchange for NAT traversal
	SIG_TYPE_ANSWER           // SDP answer in response to an offer
	SIG_TYPE_OFFER            // SDP offer to initiate connection
)
