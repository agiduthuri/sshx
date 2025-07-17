// Package conf provides configuration management for the sshx application
// It handles reading, writing, and watching configuration files using Viper
// and manages WebRTC, SSH, VNC, and other service configurations
package conf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/suutaku/go-vnc/pkg/config"
	"github.com/suutaku/sshx/internal/utils"
)

// Configure holds all configuration settings for the sshx application
// This structure is serialized to/from JSON configuration files
type Configure struct {
	// LocalSSHPort is the port where the local SSH daemon is listening (default: 22)
	LocalSSHPort int32
	
	// LocalHTTPPort is the port for HTTP services like VNC web interface (default: 80)
	LocalHTTPPort int32
	
	// LocalTCPPort is the port where sshx daemon listens for local connections (default: 2224)
	LocalTCPPort int32
	
	// ID is the unique identifier for this sshx node (UUID)
	ID string
	
	// SignalingServerAddr is the URL of the WebRTC signaling server
	SignalingServerAddr string
	
	// RTCConf contains WebRTC configuration including ICE servers for NAT traversal
	RTCConf webrtc.Configuration
	
	// VNCConf contains VNC server configuration settings
	VNCConf config.Configure
	
	// VNCStaticPath is the filesystem path to noVNC web client files
	VNCStaticPath string
	
	// ETHAddr is the ethernet address/interface to use for networking
	ETHAddr string
}

// ConfManager manages configuration lifecycle including loading, saving, and watching
// for changes to configuration files
type ConfManager struct {
	// Conf is the current configuration loaded from file
	Conf *Configure
	
	// Viper is the configuration management library instance
	Viper *viper.Viper
	
	// Path is the directory where configuration files are stored
	Path string
}

// defaultConfig provides the default configuration values for new installations
// This configuration is used when no existing config file is found
var defaultConfig = Configure{
	// Standard HTTP port for web services (VNC interface)
	LocalHTTPPort: 80,
	
	// Standard SSH port for local SSH daemon
	LocalSSHPort: 22,
	
	// Default sshx daemon listening port
	LocalTCPPort: 2224,
	
	// Generate unique identifier for this node
	ID: uuid.New().String(),
	
	// Default signaling server (should be changed in production)
	SignalingServerAddr: "http://140.179.153.231:11095",
	
	// WebRTC configuration with Google's public STUN servers
	// STUN servers help with NAT traversal by discovering public IP addresses
	RTCConf: webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				// Google's public STUN servers for NAT traversal
				URLs: []string{
					"stun:stun.l.google.com:19302",
					"stun:stun1.l.google.com:19302",
					"stun:stun2.l.google.com:19302",
					"stun:stun3.l.google.com:19302",
					"stun:stun4.l.google.com:19302",
				},
			},
		},
	},
	
	// Use default VNC configuration from the VNC library
	VNCConf: config.DefaultConfigure,
}

// ClearKnownHosts removes entries from SSH known_hosts file matching the given substring
// This prevents SSH host key verification issues when connecting to local sshx instances
// The function handles IPv4 localhost addresses by wrapping them in brackets
func ClearKnownHosts(subStr string) {
	// Convert localhost IP to bracketed format for SSH known_hosts
	// SSH uses [127.0.0.1]:port format for non-standard ports
	subStr = strings.Replace(subStr, "127.0.0.1", "[127.0.0.1]", 1)
	
	// Path to user's SSH known_hosts file
	fileName := os.Getenv("HOME") + "/.ssh/known_hosts"
	
	// Read the current known_hosts file
	input, err := ioutil.ReadFile(fileName)
	if err != nil {
		logrus.Error(err)
		return
	}
	
	// Split into lines and filter out matching entries
	lines := strings.Split(string(input), "\n")
	var newLines []string
	for i, line := range lines {
		// Skip lines that contain the substring we want to remove
		if strings.Contains(line, subStr) {
			// Skip this line (remove it)
		} else {
			// Keep this line
			newLines = append(newLines, lines[i])
		}
	}
	
	// Write the filtered content back to the file
	output := strings.Join(newLines, "\n")
	err = ioutil.WriteFile(fileName, []byte(output), 0777)
	if err != nil {
		logrus.Error(err)
		return
	}
}

// NewConfManager creates a new configuration manager instance
// It initializes Viper, loads configuration from file, and sets up file watching
// If no config file exists, it creates one with default values
func NewConfManager(homePath string) *ConfManager {
	// Use default home path if none provided
	if homePath == "" {
		homePath = utils.GetSSHXHome()
	}
	
	// Temporary configuration holder
	var tmp Configure
	
	// Initialize Viper for configuration management
	vp := viper.New()
	vp.SetConfigName(".sshx_config")    // Config file name (without extension)
	vp.SetConfigType("json")            // Configuration file format
	vp.AddConfigPath(homePath)          // Directory to search for config file
	
	// Set up configuration file watching for live reloading
	vp.WatchConfig()
	vp.OnConfigChange(func(e fsnotify.Event) {
		// Reload configuration when file changes
		err := vp.Unmarshal(&tmp)
		if err != nil {
			logrus.Error(err)
			return
		}
	})
	
	// Try to read existing configuration file
	err := vp.ReadInConfig()
	if err != nil {
		// Check if error is due to missing config file
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found - create default configuration
			
			// Generate unique peer identity for WebRTC
			defaultConfig.RTCConf.PeerIdentity = utils.HashString(fmt.Sprintf("%s%d", defaultConfig.ID, time.Now().Unix()))
			
			// Set VNC static files path
			defaultConfig.VNCStaticPath = path.Join(homePath, "noVNC")
			
			// Serialize default config to JSON
			bs, _ := json.MarshalIndent(defaultConfig, "", "  ")
			
			// Load default config into Viper
			vp.ReadConfig(bytes.NewBuffer(bs))
			
			// Write default config to file
			err = vp.WriteConfigAs(path.Join(homePath, "./.sshx_config.json"))
			if err != nil {
				logrus.Error(err)
				os.Exit(1)
			}
			
			// Make config file readable/writable
			os.Chmod(path.Join(homePath, "./.sshx_config.json"), 0777)
		} else {
			// Other error reading config file
			logrus.Error(err)
			os.Exit(1)
		}
	}

	// Unmarshal configuration into struct
	err = vp.Unmarshal(&tmp)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	// Clean up SSH known_hosts to prevent host key conflicts
	ClearKnownHosts(fmt.Sprintf("127.0.0.1:%d", tmp.LocalSSHPort))
	
	// Return initialized configuration manager
	return &ConfManager{
		Conf:  &tmp,
		Viper: vp,
		Path:  homePath,
	}
}

// Set updates a configuration value by key and persists it to the config file
// This method allows runtime configuration changes that are saved permanently
func (cm *ConfManager) Set(key, value string) {
	logrus.Info("key/value", key, value)
	
	// Update the value in Viper
	cm.Viper.Set(key, value)
	
	// Unmarshal updated config back to struct
	err := cm.Viper.Unmarshal(cm.Conf)
	if err != nil {
		logrus.Error(err)
		return
	}
	
	// Persist changes to configuration file
	err = cm.Viper.WriteConfig()
	if err != nil {
		logrus.Error(err)
		return
	}
}

// Show displays the current configuration in a formatted JSON output
// This is useful for debugging and verifying configuration settings
func (cm *ConfManager) Show() {
	// Marshal configuration to pretty-printed JSON
	bs, _ := json.MarshalIndent(cm.Conf, "", "  ")
	
	// Display configuration file location and contents
	logrus.Info("read configure file at: ", cm.Path+"/.sshx_config.json")
	logrus.Info(string(bs))
}
