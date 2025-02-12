// Auth logic

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

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

	// Secret to create authentication tokens
	AuthSecret string

	// Authentication callback URL
	AuthCallbackUrl string

	// Authentication callback authorization header
	AuthCallbackAuthorization string
}

// Authentication manager
type AuthManager struct {
	// Configuration
	config AuthConfig

	// Logger
	logger *glog.Logger

	// Users mapped to their key
	users map[string][]byte

	callbackUrl *url.URL
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

	var callbackUrl *url.URL = nil

	if config.AuthCallbackUrl != "" {
		u, err := url.Parse(config.AuthCallbackUrl)

		if err != nil {
			logger.Errorf("Invalid callback URL: %v", config.AuthCallbackUrl)
		} else {
			callbackUrl = u
		}
	}

	return &AuthManager{
		logger:      logger,
		config:      config,
		users:       users,
		callbackUrl: callbackUrl,
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
		m.logger.Debugf("[Auth request] Username not found in the list of users: %v", username)
	}

	// Auth tokens

	if len(m.config.AuthSecret) == 0 {
		// Auth tokens disabled
		m.logger.Debug("[Auth request] Authentication tokens are disabled")
		return nil, false
	}

	valid, expired, uid := ParseUsername(username)

	if !valid {
		m.logger.Debugf("[Auth request] Invalid username: %v", username)
		return nil, false
	}

	if expired {
		m.logger.Debugf("[Auth request] Expired: %v", username)
		return nil, false
	}

	// Callback

	allowed := m.MakeCallback(uid, srcAddr)

	if !allowed {
		m.logger.Debugf("[Auth request] Not allowed (Callback) / UID: %v | IP: %v", uid, srcAddr)
		return nil, false
	}

	// Generate the password

	password := GenerateAuthPassword(username, m.config.AuthSecret)

	m.logger.Debugf("[Auth request] Generated password for: %v = %v", username, password)

	return turn.GenerateAuthKey(username, m.config.Realm, password), true
}

// Parses username
func ParseUsername(username string) (valid bool, expired bool, uid string) {
	usernameParts := strings.Split(username, "/")

	if len(usernameParts) < 4 {
		return false, false, ""
	}

	userId := strings.Join(usernameParts[3:], "/")

	if strings.ToLower(usernameParts[0]) != "turn" {
		return false, false, userId
	}

	exp, err := strconv.ParseInt(usernameParts[2], 10, 64)

	if err != nil {
		return false, false, userId
	}

	now := time.Now().Unix()

	return true, now > exp, userId
}

// Generates an authentication token, to be used
// as the password for the given username
//
// Parameters:
//   - username - The username
//   - secret - The secret shared between the TURN server and the application server
//
// Returns the password as string
func GenerateAuthPassword(username string, secret string) string {
	h := sha256.New()

	h.Write([]byte(username))
	h.Write([]byte(secret))

	return strings.ToLower(hex.EncodeToString(h.Sum(nil)))
}

// Makes the authentication callback
// Requires the uid and the client IP
// Returns true only if authorized
func (m *AuthManager) MakeCallback(uid string, clientIp net.Addr) bool {
	if m.callbackUrl == nil {
		return true
	}

	finalUrl := *m.callbackUrl // Copy URL value

	finalUrl.Query().Set("uid", uid)
	finalUrl.Query().Set("ip", clientIp.String())

	finalUrlString := finalUrl.String()

	m.logger.Debugf("Calling authentication callback: GET %v", finalUrlString)

	client := &http.Client{}

	req, e := http.NewRequest("GET", finalUrlString, nil)

	if e != nil {
		m.logger.Errorf("Error creating GET request to %v - %v", finalUrlString, e.Error())
		return false
	}

	if m.config.AuthCallbackAuthorization != "" {
		req.Header.Set("Authorization", m.config.AuthCallbackAuthorization)
	}

	res, e := client.Do(req)

	if e != nil {
		m.logger.Errorf("Error sending GET request to %v - %v", finalUrlString, e.Error())
		return false
	}

	if res.StatusCode != 200 {
		m.logger.Debugf("Status code: %v as a response of GET %v", res.StatusCode, finalUrlString)
		return false
	}

	return true
}
