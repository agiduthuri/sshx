// Package types - signal.go defines structures for WebRTC signaling messages
// These messages are exchanged between peers through the signaling server
package types

// SignalingInfo represents a WebRTC signaling message exchanged between peers
// This structure carries all the information needed for WebRTC peer discovery,
// SDP offer/answer exchange, and ICE candidate sharing
type SignalingInfo struct {
	// Flag indicates the type of signaling message (SIG_TYPE_OFFER, SIG_TYPE_ANSWER, etc.)
	Flag int `json:"flag"`
	
	// Source is the peer ID of the sender
	Source string `json:"source"`
	
	// SDP contains the Session Description Protocol data for WebRTC handshake
	// This includes codec information, network details, and session parameters
	SDP string `json:"sdp"`
	
	// Candidate contains serialized ICE candidate data for NAT traversal
	// ICE candidates provide network path information for establishing connections
	Candidate []byte `json:"candidate"`
	
	// Id is the unique identifier for the connection pool/session
	Id PoolId `json:"id"`
	
	// Target is the peer ID of the intended recipient
	Target string `json:"target"`
	
	// PeerType indicates the role of the peer (dialer/responser)
	PeerType int32 `json:"peer_type"`
	
	// RemoteRequestType specifies the type of application/service being requested
	// (APP_TYPE_SSH, APP_TYPE_VNC, etc.)
	RemoteRequestType int32 `json:"remote_request_type"`
}
