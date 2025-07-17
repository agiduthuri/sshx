# SSHX Project Documentation

## Overview

SSHX is a comprehensive WebRTC-based SSH remote toolbox written in Go that provides secure remote access and file transfer capabilities through peer-to-peer connections. The project enables SSH connections, file transfers, VNC remote desktop access, and various other services through both direct TCP connections and WebRTC peer-to-peer connections with NAT traversal.

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                                    SSHX System                                             │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐           ┌─────────────────┐           ┌─────────────────┐       │
│  │   CLI Commands  │           │ Signaling Server │           │   Node Services │       │
│  │                 │           │                 │           │                 │       │
│  │ • daemon        │           │ • Pull/Push     │           │ • TCP Server    │       │
│  │ • connect       │           │ • Peer Discovery│           │ • Connection    │       │
│  │ • copy          │           │ • Message Queue │           │   Management    │       │
│  │ • proxy         │           │ • Watchdog      │           │ • Configuration │       │
│  │ • status        │           │                 │           │                 │       │
│  │ • etc.          │           │                 │           │                 │       │
│  └─────────────────┘           └─────────────────┘           └─────────────────┘       │
│           │                            │                            │                  │
│           │                            │                            │                  │
├───────────┼────────────────────────────┼────────────────────────────┼──────────────────┤
│           │                            │                            │                  │
│  ┌─────────────────────────────────────────────────────────────────────────────────────┐ │
│  │                        Connection Management Layer                                    │ │
│  │                                                                                       │ │
│  │  ┌─────────────────┐                            ┌─────────────────┐                │ │
│  │  │ Direct Service  │                            │ WebRTC Service  │                │ │
│  │  │                 │                            │                 │                │ │
│  │  │ • TCP Direct    │                            │ • P2P WebRTC    │                │ │
│  │  │ • Local Network │                            │ • NAT Traversal │                │ │
│  │  │ • Fast          │                            │ • Signaling     │                │ │
│  │  │                 │                            │ • ICE/STUN      │                │ │
│  │  └─────────────────┘                            └─────────────────┘                │ │
│  └─────────────────────────────────────────────────────────────────────────────────────┘ │
│                                          │                                                │
├──────────────────────────────────────────┼────────────────────────────────────────────────┤
│                                          │                                                │
│  ┌─────────────────────────────────────────────────────────────────────────────────────┐ │
│  │                        Application Implementation Layer                               │ │
│  │                                                                                       │ │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐      │ │
│  │  │   SSH   │  │   SCP   │  │   VNC   │  │  PROXY  │  │  SSHFS  │  │   STAT  │      │ │
│  │  │         │  │         │  │         │  │         │  │         │  │         │      │ │
│  │  │ Remote  │  │ File    │  │ Desktop │  │ Network │  │ File    │  │ Monitor │      │ │
│  │  │ Shell   │  │ Transfer│  │ Access  │  │ Proxy   │  │ System  │  │ Status  │      │ │
│  │  │         │  │         │  │         │  │         │  │         │  │         │      │ │
│  │  └─────────┘  └─────────┘  └─────────┘  └─────────┘  └─────────┘  └─────────┘      │ │
│  │                                                                                       │ │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐                                │ │
│  │  │ MESSAGE │  │TRANSFER │  │TRANSFER │  │   VNC   │                                │ │
│  │  │         │  │ CLIENT  │  │ SERVICE │  │ SERVICE │                                │ │
│  │  │ Console │  │         │  │         │  │         │                                │ │
│  │  │         │  │         │  │         │  │         │                                │ │
│  │  └─────────┘  └─────────┘  └─────────┘  └─────────┘                                │ │
│  └─────────────────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Command Layer (`cmd/`)

#### Main Client (`cmd/sshx/`)
- **Entry Point**: `main.go` - CLI application with command routing
- **Commands**: 
  - `daemon` - Start the main node service
  - `connect` - SSH connection to remote hosts
  - `copy` - Secure file copy operations
  - `proxy` - HTTP/SOCKS proxy functionality
  - `fs` - SSH filesystem mounting
  - `vnc` - VNC remote desktop access
  - `status` - Connection monitoring
  - `transfer` - File transfer operations
  - `message` - Real-time messaging

#### Signaling Server (`cmd/signaling/`)
- **Purpose**: WebRTC signaling server for peer discovery
- **Protocol**: HTTP-based pull/push messaging
- **Features**:
  - Peer message queuing with 64-message buffer
  - Automatic peer cleanup after 15 seconds of inactivity
  - Watchdog mechanism for resource management
  - Binary message encoding using Go's `gob` format

### 2. Node Management (`internal/node/`)

The Node is the central coordination component that:
- Manages configuration through `ConfManager`
- Coordinates connection services (Direct TCP and WebRTC)
- Provides TCP server interface for local connections
- Handles service lifecycle (start/stop)

### 3. Connection Management (`internal/conn/`)

#### Connection Services
- **Direct Service**: Fast TCP connections for local networks
- **WebRTC Service**: P2P connections with NAT traversal for remote access

#### Connection Types
- **Direct Connections**: Standard TCP connections
- **WebRTC Connections**: Peer-to-peer connections using WebRTC data channels

#### Connection Manager
- Manages multiple connection services
- Handles connection lifecycle and state management
- Provides connection pooling and resource management

### 4. Application Implementation Layer (`pkg/impl/`)

#### Base Implementation (`impl_base.go`)
- Common functionality for all application types
- Connection management and peer identification
- Thread-safe I/O operations

#### Application Types
1. **SSH** (`impl_ssh.go`) - Remote shell access
2. **SCP** (`impl_scp.go`) - Secure file copy
3. **VNC** (`impl_vnc.go`) - VNC client for remote desktop
4. **VNC Service** (`impl_vnc_service.go`) - VNC server
5. **Proxy** (`impl_proxy.go`) - HTTP/SOCKS proxy server
6. **SSHFS** (`impl_sshfs.go`) - SSH filesystem mounting
7. **Transfer** (`impl_transfer.go`) - File transfer client
8. **Transfer Service** (`impl_transfer_service.go`) - File transfer server
9. **Messager** (`impl_messager.go`) - Real-time messaging
10. **Status** (`impl_stat.go`) - Connection monitoring

### 5. Configuration Management (`pkg/conf/`)

#### Configuration Structure
- **Ports**: SSH (22), HTTP (80), TCP (2224)
- **WebRTC**: ICE servers, peer identity, signaling server
- **VNC**: Desktop sharing configuration
- **Paths**: Static files, configuration directory

#### Features
- JSON configuration files with live reloading
- Default configuration generation
- SSH known_hosts management
- Viper integration for configuration management

### 6. Type System (`pkg/types/`)

#### Core Types
- **Application Types**: Constants for different services
- **Option Types**: Connection operation types (UP, DOWN, STAT, ATTACH)
- **Signaling Types**: WebRTC message types (OFFER, ANSWER, CANDIDATE)

#### Data Structures
- **SignalingInfo**: WebRTC signaling messages
- **PoolId**: Connection pool identification
- **Status**: Connection and system status

## WebRTC Signaling Protocol

### Message Flow
```
Node A (Dialer)                    Signaling Server                    Node B (Responser)
      │                                   │                                   │
      │────── POST /push/{target_id} ────▶│                                   │
      │       (SDP Offer)                 │                                   │
      │                                   │◀──── GET /pull/{self_id} ────────│
      │                                   │       (Poll for messages)         │
      │                                   │                                   │
      │                                   │────── SDP Offer ────────────────▶│
      │                                   │                                   │
      │                                   │◀──── POST /push/{target_id} ─────│
      │                                   │       (SDP Answer)                │
      │◀──── GET /pull/{self_id} ────────│                                   │
      │       (Poll for messages)         │                                   │
      │                                   │                                   │
      │◀────── SDP Answer ───────────────│                                   │
      │                                   │                                   │
      │                                   │                                   │
      │◀─────────────── WebRTC Data Channel ──────────────────────────────▶│
      │                (Direct P2P Connection)                               │
```

### Signaling Message Structure
```go
type SignalingInfo struct {
    Flag              int    // Message type (OFFER, ANSWER, CANDIDATE)
    Source            string // Sender peer ID
    SDP               string // Session Description Protocol data
    Candidate         []byte // ICE candidate information
    Id                PoolId // Connection pool identifier
    Target            string // Recipient peer ID
    PeerType          int32  // Peer role (dialer/responser)
    RemoteRequestType int32  // Application type being requested
}
```

## Connection Establishment Process

### 1. Local Connection
```
Client Application → TCP :2224 → Node → Connection Manager → Application Implementation
```

### 2. Direct TCP Connection
```
Node A → Direct Service → TCP Connection → Node B → Application Implementation
```

### 3. WebRTC P2P Connection
```
Node A → WebRTC Service → Signaling Server → Node B → WebRTC Service
                     ↓
              WebRTC Data Channel (P2P)
                     ↓
            Application Implementation
```

## Application Implementation Interface

All applications implement the `Impl` interface:

```go
type Impl interface {
    Init()                          // Initialize the implementation
    Code() int32                    // Return application type code
    SetConn(net.Conn)              // Set network connection
    Conn() net.Conn                // Get network connection
    Writer() io.Writer             // Get writer for sending data
    Reader() io.Reader             // Get reader for receiving data
    Response() error               // Handle incoming requests
    Dial() error                   // Initiate outgoing connections
    Preper() error                 // Prepare for connection
    Close()                        // Clean up resources
    SetHostId(string)              // Set host identifier
    HostId() string                // Get host identifier
    PairId() string                // Get peer identifier
    SetPairId(string)              // Set peer identifier
    ParentId() string              // Get parent connection ID
    SetParentId(string)            // Set parent connection ID
    Attach(net.Conn) error         // Attach to existing connection
    NoNeedConnect()                // Mark as not needing connection
    IsNeedConnect() bool           // Check if connection needed
}
```

## Security Features

### 1. SSH Key Management
- Public key authentication support
- SSH known_hosts management
- Automatic host key cleanup

### 2. WebRTC Security
- ICE/STUN for NAT traversal
- Peer identity verification
- Encrypted data channels

### 3. Configuration Security
- Secure configuration file handling
- Environment variable overrides
- Path validation

## Build and Deployment

### Build System
- **Go Modules**: Dependency management
- **Build Script**: `build.sh` for compilation and installation
- **Makefile**: Simplified build commands
- **System Integration**: Service files for Linux/macOS

### Dependencies
- **WebRTC**: `github.com/pion/webrtc/v3`
- **CLI**: `github.com/jawher/mow.cli`
- **Configuration**: `github.com/spf13/viper`
- **Logging**: `github.com/sirupsen/logrus`
- **Networking**: Standard Go networking packages

### Installation
```bash
# Build both components
./build.sh

# Install as system service
sudo ./build.sh install

# Install with signaling server
sudo ./build.sh install signaling
```

## Usage Examples

### Basic SSH Connection
```bash
# Start daemon
sshx daemon

# Connect to remote host
sshx connect user@target-host-id

# Copy files
sshx copy file.txt user@target-host-id:/path/

# Mount filesystem
sshx fs mount user@target-host-id:/path/ /local/mount/
```

### VNC Remote Desktop
```bash
# Start VNC service
sshx vnc service start

# Connect to VNC (web interface)
# Access http://127.0.0.1 or http://vnc.sshx.wz
```

### Proxy Usage
```bash
# Start proxy server
sshx proxy start

# Use with applications requiring proxy
```

## Configuration

### Default Configuration
```json
{
  "LocalSSHPort": 22,
  "LocalHTTPPort": 80,
  "LocalTCPPort": 2224,
  "ID": "unique-node-id",
  "SignalingServerAddr": "http://signaling-server:11095",
  "RTCConf": {
    "iceServers": [
      {
        "urls": ["stun:stun.l.google.com:19302"]
      }
    ]
  }
}
```

### Environment Variables
- `SSHX_HOME`: Override default configuration directory
- `SSHX_SIGNALING_PORT`: Signaling server port (default: 11095)

## Troubleshooting

### Common Issues
1. **Connection Failures**: Check signaling server connectivity
2. **NAT Traversal**: Verify STUN server configuration
3. **Port Conflicts**: Adjust port configurations
4. **SSH Key Issues**: Check SSH key permissions and known_hosts

### Debugging
- Enable debug logging with environment variables
- Check configuration file syntax
- Verify network connectivity
- Monitor connection status

## Future Enhancements

### Missing Features
- **Unit Tests**: No test coverage currently exists
- **Documentation**: Limited inline documentation
- **Monitoring**: Basic status monitoring only
- **Authentication**: Basic SSH key authentication

### Potential Improvements
- Enhanced error handling
- Performance optimizations
- Better logging and metrics
- More robust peer discovery
- Advanced security features

## Conclusion

SSHX provides a comprehensive solution for remote access and file transfer with both traditional TCP and modern WebRTC peer-to-peer capabilities. The modular architecture allows for easy extension with new application types while maintaining a clean separation of concerns between connection management, signaling, and application logic.

The project demonstrates effective use of Go's concurrency features, WebRTC technology, and system integration capabilities to create a powerful remote access toolbox suitable for various network environments and use cases.