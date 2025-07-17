// Package impl - impl_base.go provides the base implementation for all application types
// This base class contains common functionality shared by all application implementations
package impl

import (
	"io"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
)

// BaseImpl provides common functionality for all application implementations
// It handles connection management, peer identification, and basic I/O operations
// All specific implementations (SSH, VNC, SCP, etc.) embed this base structure
type BaseImpl struct {
	// HId is the host identifier for this implementation instance
	HId string
	
	// conn is the network connection used for communication
	conn *net.Conn
	
	// Parent is the ID of the parent connection or session
	Parent string
	
	// PId is the peer identifier for the remote end
	PId string
	
	// lock provides thread-safe access to the connection
	lock sync.Mutex
	
	// ConnectNow indicates whether this implementation needs an active connection
	ConnectNow bool
}

func NewBaseImpl(hid string) *BaseImpl {
	return &BaseImpl{
		ConnectNow: true,
		HId:        hid,
	}
}

func (base *BaseImpl) IsNeedConnect() bool {
	return base.ConnectNow
}

func (base *BaseImpl) NoNeedConnect() {
	base.ConnectNow = false
}

func (base *BaseImpl) Init() {}

func (base *BaseImpl) Conn() net.Conn {
	base.lock.Lock()
	defer base.lock.Unlock()
	return *base.conn
}

func (base *BaseImpl) Preper() error {
	return nil
}

func (base *BaseImpl) PairId() string {
	return base.PId
}

func (base *BaseImpl) SetHostId(id string) {
	if id == "" {
		logrus.Warn("Set empty string to host id")
	}
	base.HId = id
}

func (base *BaseImpl) SetPairId(id string) {
	base.PId = id
}

func (base *BaseImpl) ParentId() string {
	return base.Parent
}

func (base *BaseImpl) SetParentId(id string) {
	base.Parent = id
}

func (base *BaseImpl) SetConn(conn net.Conn) {
	base.lock.Lock()
	defer base.lock.Unlock()
	logrus.Debug("set connection (non-detach)")
	base.conn = &conn
}

func (base *BaseImpl) Reader() io.Reader {
	base.lock.Lock()
	defer base.lock.Unlock()
	return *(base.conn)
}

func (base *BaseImpl) Writer() io.Writer {
	base.lock.Lock()
	defer base.lock.Unlock()
	return (*base.conn)
}

func (base *BaseImpl) ReadWriteCloser() io.ReadWriteCloser {
	base.lock.Lock()
	defer base.lock.Unlock()
	return (*base.conn)
}

func (base *BaseImpl) HostId() string {
	return base.HId
}

func (base *BaseImpl) Close() {
	if base.conn != nil {
		logrus.Debug("close Conn")
		(*base.conn).Close()
	}
	logrus.Debug("close base impl")
}

// Response of remote device call
func (base *BaseImpl) Response() error {
	return nil
}

// Call remote device
func (base *BaseImpl) Dial() error {
	return nil
}

func (base *BaseImpl) Attach(net.Conn) error {
	return nil
}
