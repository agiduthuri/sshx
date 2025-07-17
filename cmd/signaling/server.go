package main

import (
	"encoding/gob"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/suutaku/sshx/pkg/types"
)

// Server represents the WebRTC signaling server
// This server facilitates peer discovery and SDP exchange for WebRTC connections
// It uses HTTP endpoints for peers to exchange offers/answers and ICE candidates
type Server struct {
	port string    // Port to listen on for HTTP requests
	dm   *DManager // Data manager handles peer message queues and lifecycle
}

// NewServer creates a new signaling server instance
// port: The port number to bind the HTTP server to
func NewServer(port string) *Server {
	return &Server{
		port: port,
		dm:   NewDManager(), // Initialize data manager for peer messaging
	}
}

// Start launches the HTTP server with routing endpoints
// Sets up two main routes:
// - /pull/{self_id}: Endpoint for peers to retrieve messages destined for them
// - /push/{target_id}: Endpoint for peers to send messages to other peers
func (sv *Server) Start() {
	// Create HTTP router using gorilla/mux
	r := mux.NewRouter()
	
	// Route for peers to pull messages addressed to them
	// self_id is the ID of the peer requesting messages
	r.Handle("/pull/{self_id}", sv.pull())
	
	// Route for peers to push messages to other peers
	// target_id is the ID of the peer to receive the message
	r.Handle("/push/{target_id}", sv.push())

	// Register router with default HTTP handler
	http.Handle("/", r)

	// Start HTTP server - this blocks until server stops
	logrus.Infof("Listening on port %s", sv.port)
	logrus.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", sv.port), nil))
}

// pull handles HTTP requests from peers wanting to retrieve messages
// This implements a polling mechanism where peers periodically check for new messages
// Uses non-blocking channel read to avoid hanging if no messages are available
func (sv *Server) pull() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		self_id := vars["self_id"] // Extract peer ID from URL path
		
		// Non-blocking read from peer's message channel
		select {
		case v := <-sv.dm.Get(self_id):
			// Message available - encode and send it
			logrus.Debug("pull from ", self_id, v.Flag)
			w.Header().Add("Content-Type", "application/binary")
			
			// Encode SignalingInfo as binary using gob
			if err := gob.NewEncoder(w).Encode(v); err != nil {
				logrus.Error("binary encode failed:", err)
				return
			}
		default:
			// No messages available - return empty response
			// Client will poll again later
		}
	})
}

// push handles HTTP requests from peers wanting to send messages to other peers
// Receives SignalingInfo (SDP offers/answers, ICE candidates) and queues them
// for the target peer to retrieve via pull requests
func (sv *Server) push() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var info types.SignalingInfo
		
		// Decode binary SignalingInfo from request body
		if err := gob.NewDecoder(r.Body).Decode(&info); err != nil {
			logrus.Error("binary decode failed:", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		
		vars := mux.Vars(r)
		target_id := vars["target_id"] // Extract target peer ID from URL path
		
		// Queue message for target peer and reset their keepalive timer
		sv.dm.Set(target_id, info)
		logrus.Debug("push from ", info.Source, " to ", target_id, info.Flag)
	})
}
