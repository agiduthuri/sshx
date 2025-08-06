# SSHX Implementation Framework Reading Guide

## Overview

This directory contains the implementation framework for SSHX's plugin-style architecture. All services (SSH, SCP, VNC, etc.) implement a common `Impl` interface, enabling a unified connection management system over WebRTC.

The framework follows a modular design where each service is a separate implementation that inherits from a common base class, providing consistency while allowing service-specific customizations.

## Architecture Overview

### Core Interface

All implementations must satisfy the `Impl` interface defined in `impl.go:14-43`:

```go
type Impl interface {
    Init()
    Code() int32                    // Application type identifier
    SetConn(net.Conn) / Conn()     // Connection management
    Writer() / Reader()            // I/O streams
    Response() error               // Handle incoming connections
    Dial() error                   // Initiate outgoing connections
    Preper() error                 // Preparation before connection
    Close()                        // Cleanup
    // ID management
    SetHostId(string) / HostId()
    SetPairId(string) / PairId()
    SetParentId(string) / ParentId()
    Attach(net.Conn) error         // Attach to existing connection
    NoNeedConnect() / IsNeedConnect() // Connection control
}
```

### Key Components

- **BaseImpl**: Common implementation providing thread-safe connection management
- **Sender**: Handles serialization and communication with local daemon
- **Registry**: Dynamic service discovery and instantiation

## Reading Order (Dependency-Based)

### 1. Core Framework Files 📚

Start with these files to understand the fundamental architecture:

#### `impl.go` - The Foundation
- **Lines of Interest**: 14-43 (interface), 45-56 (registry), 58-66 (factory)
- **Purpose**: Defines the core `Impl` interface that all services implement
- **Key Features**:
  - Registry of all applications (`registeddApp` slice)
  - Factory function `GetImpl()` for dynamic instantiation via reflection
  - Service discovery by application code

#### `impl_base.go` - Common Implementation
- **Purpose**: Base class that all services inherit from
- **Key Features**:
  - Thread-safe connection management with mutex
  - Default implementations for most interface methods
  - ID management (host, pair, parent IDs)
  - Standard I/O stream handling

#### `sender.go` - Communication Protocol
- **Purpose**: Handles encoding and sending requests to local daemon
- **Key Features**:
  - `Sender` struct for request serialization
  - Uses `gob` encoding for data transmission
  - Type field combines app code and option code via bit shifting
  - TCP communication with local daemon

### 2. Core Service Implementations 🔧

#### `impl_ssh.go` (355 lines) - SSH Terminal Service
- **Complexity**: High - Most comprehensive implementation
- **Key Methods**:
  - `OpenTerminal()`: SSH session setup and terminal handling (lines 121-196)
  - `passwordCallback()`: Interactive authentication (lines 199-206)
  - `hostKeyCallback()`: Known hosts management (lines 229-256)
- **Features**:
  - Full SSH client functionality
  - X11 forwarding support (`x11Request()`, `forwardX11Socket()`)
  - Public key and password authentication
  - Terminal size and mode handling
  - Known hosts file management

#### `impl_proxy.go` - Local Proxy Service
- **Purpose**: Creates local TCP proxy for SSH connections
- **Connection Type**: Detached (background service)
- **Key Method**: `doDial()` - handles incoming connections and pipes through WebRTC
- **Use Case**: Allows standard SSH clients to connect through SSHX tunnels

#### `impl_scp.go` - Secure Copy Functionality
- **Purpose**: File transfer using SSH protocol
- **Key Methods**:
  - `ParsePaths()`: Determines transfer direction from source/dest paths
  - `Dial()`: Establishes SSH connection and performs transfer
- **Dependencies**: Uses SSH implementation for authentication

### 3. Advanced Services 🚀

#### `impl_sshfs.go` - SSH Filesystem Mounting
- **Purpose**: SFTP-based filesystem using FUSE
- **Dependencies**: Relies on SSH implementation for connection setup
- **Key Features**:
  - Mount/unmount operations
  - SFTP client integration
  - Debug mode support

#### `impl_vnc.go` & `impl_vnc_service.go` - VNC Remote Desktop
- **`impl_vnc.go`**: Client-side VNC connection via WebSocket
  - Connects to local websockify service
  - Upgrades WebSocket to raw TCP connection
- **`impl_vnc_service.go`**: Server-side HTTP service
  - Web-based VNC viewer with WebSocket upgrade
  - HTTP server with static file serving
  - Real-time VNC protocol handling

### 4. File Transfer Services 📁

#### `impl_transfer.go` - Direct File Transfer
- **Protocol**: Header exchange followed by binary data stream
- **Key Methods**:
  - `sendHeader()`/`recvHeader()`: File metadata exchange
  - `DoUpload()`/`DoDownload()`: Transfer operations with progress bars
- **Features**:
  - Progress bar integration using `progressbar/v3`
  - Support for both file paths and `io.Reader`/`io.Writer`
  - Automatic file creation in Downloads folder

#### `impl_transfer_service.go` - HTTP-based File Transfer
- **Purpose**: Web service with QR code generation for easy mobile access
- **Key Features**:
  - HTTP endpoints for upload/download
  - QR code generation for mobile scanning
  - Temporary file caching
  - Brotli-compressed web interface
  - Random URL generation for security

### 5. Communication & Utility Services 💬

#### `impl_messager.go` - P2P Messaging
- **Architecture**: Channel-based message queuing with goroutines
- **Key Features**:
  - Terminal UI for interactive chat (`OpenChatConsole()`)
  - Desktop notifications when UI not active
  - Concurrent send/receive operations
  - Message buffering with channel overflow handling
- **Threading**: Separate goroutines for send and receive operations

#### `impl_stat.go` - Status Monitoring
- **Purpose**: Display active connections and service status
- **Display Modes**:
  - Table format (`showTable()`)
  - Tree format (`showList()`) - shows parent/child relationships
- **Data Exchange**: Uses `gob` encoding for status information

## Key Architectural Patterns

### Connection Types
- **Attached**: Direct I/O connection (SSH terminal, VNC viewer)
  - Connection directly handles user interaction
  - Typically synchronous operation
- **Detached**: Background services (Proxy, Messager, Transfer Service)
  - Service runs independently
  - Asynchronous operation with event handling

### Threading Model
- **BaseImpl**: Provides mutex-based thread safety for connection access
- **Service-specific**: Most implementations use goroutines for concurrent I/O
- **Cleanup**: Services handle their own resource cleanup in `Close()` methods

### Error Handling Strategy
- **Logging**: Extensive use of `logrus` for debugging and error tracking
- **Propagation**: Errors returned through standard Go error interface
- **Graceful Degradation**: Services attempt to continue operation when possible
- **Resource Cleanup**: Defer statements ensure proper resource management

### Communication Flow

1. **Initialization**: Client creates implementation instance
2. **Serialization**: Implementation wrapped in `Sender` struct
3. **Transport**: Sent to local daemon via TCP connection
4. **WebRTC**: Local daemon establishes P2P connection via signaling server
5. **Remote Handling**: Remote daemon receives and creates matching implementation
6. **Protocol**: Service-specific protocol handles the established connection

## Implementation Guidelines

### Adding New Services

1. **Create Implementation**: Inherit from `BaseImpl`
2. **Implement Interface**: Satisfy all `Impl` interface methods
3. **Register Service**: Add to `registeddApp` slice in `impl.go`
4. **Define Code**: Add unique application code in `pkg/types/types.go`
5. **Handle Serialization**: Ensure struct fields are exported for `gob` encoding

### Best Practices

- **Thread Safety**: Use `BaseImpl`'s mutex for connection access
- **Resource Management**: Always implement proper cleanup in `Close()`
- **Error Handling**: Log errors with context using logrus
- **Configuration**: Use config manager for service-specific settings
- **Testing**: Ensure service works in both dial and response modes

## Dependencies

### External Libraries
- **SSH**: `golang.org/x/crypto/ssh` for SSH protocol
- **WebSocket**: `gorilla/websocket` for VNC service
- **Progress**: `schollz/progressbar/v3` for file transfers
- **UI**: `jedib0t/go-pretty/v6` for status display
- **Filesystem**: `hanwen/go-fuse/v2` for SSHFS
- **Notifications**: `martinlindhe/notify` for desktop alerts

### Internal Dependencies
- **Configuration**: `pkg/conf` for application settings
- **Types**: `pkg/types` for application codes and shared structures
- **Utilities**: `internal/utils` for helper functions

## Debugging Tips

1. **Enable Debug Mode**: Set debug flags in utils package
2. **Check Logs**: All implementations use logrus for detailed logging
3. **Connection Issues**: Verify `BaseImpl.conn` is properly set
4. **Serialization**: Ensure all struct fields are exported for gob encoding
5. **Threading**: Use mutex appropriately when accessing shared resources

This framework provides a robust foundation for implementing various network services over P2P WebRTC connections, with consistent patterns and comprehensive error handling throughout.