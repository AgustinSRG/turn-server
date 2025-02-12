// Main

package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AgustinSRG/genv"
	"github.com/AgustinSRG/glog"
	"github.com/joho/godotenv"
	"github.com/pion/turn/v4"

	tls_certificate_loader "github.com/AgustinSRG/go-tls-certificate-loader"
)

// Main function (program entry point)
func main() {
	_ = godotenv.Load() // Load env vars

	// Configure logs

	logger := glog.CreateRootLogger(glog.LoggerConfiguration{
		ErrorEnabled:   genv.GetEnvBool("LOG_ERROR", true),
		WarningEnabled: genv.GetEnvBool("LOG_WARNING", true),
		InfoEnabled:    genv.GetEnvBool("LOG_INFO", true),
		DebugEnabled:   genv.GetEnvBool("LOG_DEBUG", false),
		TraceEnabled:   genv.GetEnvBool("LOG_TRACE", false),
	}, glog.StandardLogFunction)

	loggerFactory := &LoggerWrapperFactory{
		logger: logger,
	}

	// Load configuration

	realm := genv.GetEnvString("REALM", "pion.ly")

	logger.Infof("[CONFIG] Realm: %v", realm)

	var publicIp *net.IP = nil

	publicIpStr := genv.GetEnvString("PUBLIC_IP", "")

	if publicIpStr != "" {
		ip := net.ParseIP(publicIpStr)

		if ip == nil {
			logger.Errorf("Invalid IP address (PUBLIC_IP): %v", publicIpStr)
			os.Exit(1)
		}

		publicIp = &ip
	} else {
		detectedIp, err := DetectExternalIPAddress()

		if err != nil {
			logger.Errorf("Could not detect a public IP: %v - Please set the PUBLIC_IP environment variable", err)
			os.Exit(1)
		}

		if detectedIp == nil {
			logger.Error("Could not detect a public IP - Please set the PUBLIC_IP environment variable")
			os.Exit(1)
		}

		publicIp = detectedIp
	}

	logger.Infof("[CONFIG] Public IP: %v", publicIp)

	// Create auth manager

	authManager := NewAuthManager(logger.CreateChildLogger("[AUTH] "), AuthConfig{
		Realm:                     realm,
		Users:                     genv.GetEnvString("USERS", ""),
		AuthSecret:                genv.GetEnvString("AUTH_SECRET", ""),
		AuthCallbackUrl:           genv.GetEnvString("AUTH_CALLBACK_URL", ""),
		AuthCallbackAuthorization: genv.GetEnvString("AUTH_CALLBACK_AUTHORIZATION", ""),
	})

	// Create relay address generator

	minPort := genv.GetEnvUint16("MIN_RELAY_PORT", 50000)
	maxPort := genv.GetEnvUint16("MAX_RELAY_PORT", 55000)

	relayAddressGenerator := &turn.RelayAddressGeneratorPortRange{
		RelayAddress: *publicIp,
		Address:      publicIp.String(),
		MinPort:      minPort,
		MaxPort:      maxPort,
	}

	// Create listeners

	listenerConfigs := make([]turn.ListenerConfig, 0)
	packetConnConfigs := make([]turn.PacketConnConfig, 0)

	// UDP listener

	if genv.GetEnvBool("UDP_ENABLED", true) {
		port := genv.GetEnvUint16("UDP_PORT", 3478)
		bindAddress := genv.GetEnvString("UDP_BIND_ADDRESS", "")

		listenAddr := bindAddress + ":" + fmt.Sprint(port)

		udpListener, err := net.ListenPacket("udp", listenAddr)

		if err != nil {
			logger.Errorf("Failed to create UDP listener: %s", err)
			os.Exit(1)
		}

		packetConnConfigs = append(packetConnConfigs, turn.PacketConnConfig{
			PacketConn:            udpListener,
			RelayAddressGenerator: relayAddressGenerator,
		})
		logger.Infof("Created UDP listener bound to %v", listenAddr)
	}

	// TCP listener

	if genv.GetEnvBool("TCP_ENABLED", true) {
		port := genv.GetEnvUint16("TCP_PORT", 3478)
		bindAddress := genv.GetEnvString("TCP_BIND_ADDRESS", "")

		listenAddr := bindAddress + ":" + fmt.Sprint(port)

		tcpListener, err := net.Listen("tcp", listenAddr)

		if err != nil {
			logger.Errorf("Failed to create TCP listener: %s", err)
			os.Exit(1)
		}

		listenerConfigs = append(listenerConfigs, turn.ListenerConfig{
			Listener:              tcpListener,
			RelayAddressGenerator: relayAddressGenerator,
		})

		logger.Infof("Created TCP listener bound to %v", listenAddr)
	}

	// TLS listener

	if genv.GetEnvBool("TLS_ENABLED", false) {
		port := genv.GetEnvUint16("TLS_PORT", 5349)
		bindAddress := genv.GetEnvString("TLS_BIND_ADDRESS", "")

		listenAddr := bindAddress + ":" + fmt.Sprint(port)

		tlsLoader, err := tls_certificate_loader.NewTlsCertificateLoader(tls_certificate_loader.TlsCertificateLoaderConfig{
			// Path to the certificate and the key
			CertificatePath: genv.GetEnvString("TLS_CERTIFICATE", "certificate.pem"),
			KeyPath:         genv.GetEnvString("TLS_PRIVATE_KEY", "key.pem"),

			// Interval to check for changes
			CheckReloadPeriod: time.Duration(genv.GetEnvUint("TLS_CHECK_RELOAD_SECONDS", 60)) * time.Minute,

			// Event functions
			OnReload: func() {
				logger.Info("[TLS] Certificate was reloaded!")
			},
			OnError: func(err error) {
				logger.Errorf("[TLS] Error loading certificate: %v \n", err)
			},
		})

		if err != nil {
			logger.Errorf("[TLS] Error loading certificate: %v \n", err)
			os.Exit(1)
		}

		tlsListener, err := tls.Listen("tcp", listenAddr, &tls.Config{
			GetCertificate: tlsLoader.GetCertificate,
		})

		if err != nil {
			logger.Errorf("Failed to create TLS listener: %s", err)
			os.Exit(1)
		}

		listenerConfigs = append(listenerConfigs, turn.ListenerConfig{
			Listener:              tlsListener,
			RelayAddressGenerator: relayAddressGenerator,
		})

		logger.Infof("Created TLS listener bound to %v", listenAddr)
	}

	// Create TURN server

	if len(listenerConfigs) == 0 && len(packetConnConfigs) == 0 {
		logger.Error("No listeners enabled")
		os.Exit(1)
	}

	server, err := turn.NewServer(turn.ServerConfig{
		Realm:             realm,
		AuthHandler:       authManager.HandleAuth,
		LoggerFactory:     loggerFactory,
		ListenerConfigs:   listenerConfigs,
		PacketConnConfigs: packetConnConfigs,
	})

	if err != nil {
		logger.Errorf("Could not start the server: %v", err)
		os.Exit(1)
	}

	logger.Info("Successfully started TURN server")

	// Block until user sends SIGINT or SIGTERM

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	// Close server

	if err = server.Close(); err != nil {
		logger.Errorf("Could not close the server: %v", err)
		os.Exit(1)
	}
}
