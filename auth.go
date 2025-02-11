// Auth logic

package main

import (
	"net"
	"strings"

	"github.com/AgustinSRG/glog"
	"github.com/pion/turn/v4"
)

// Authentication configuration
type AuthConfig struct {
	// TURN realm
	Realm string

	// List of users + passwords, separated by commas
	// The username and password are separated with a colon
	Users string
}

// Authentication manager
type AuthManager struct {
	// Configuration
	config AuthConfig

	// Logger
	logger *glog.Logger

	// Users mapped to their key
	users map[string][]byte
}

// Creates new AuthManager
func NewAuthManager(logger *glog.Logger, config AuthConfig) *AuthManager {
	users := make(map[string][]byte)

	if config.Users != "" {
		usersParts := strings.Split(config.Users, ",")

		for _, userPart := range usersParts {
			userSplit := strings.Split(userPart, ":")

			if len(userSplit) != 2 {
				logger.Warningf("Could not parse username:password pair: %v", userPart)
				continue
			}

			username := strings.TrimSpace(userSplit[0])
			password := strings.TrimSpace(userSplit[1])

			users[userSplit[0]] = turn.GenerateAuthKey(username, config.Realm, password)
		}
	}

	return &AuthManager{
		logger: logger,
		config: config,
		users:  users,
	}
}

// Handles authentication
func (m *AuthManager) HandleAuth(username, realm string, srcAddr net.Addr) (key []byte, ok bool) {
	if m.logger.Config.DebugEnabled {
		m.logger.Debugf("[Auth request] Username=%v, Realm=%v, IP=%v", username, realm, srcAddr)
	}

	if key, ok := m.users[username]; ok {
		return key, true
	}

	if m.logger.Config.DebugEnabled {
		m.logger.Debugf("[Auth request] Username not found: %v", username)
	}

	return nil, false
}
