///////////////////////////////////////////////////////////////////////
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////

package http

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-openapi/swag"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme"
	"golang.org/x/net/netutil"

	"github.com/vmware/dispatch/pkg/utils"
)

const (
	schemeHTTP  = "http"
	schemeHTTPS = "https"

	defaultCleanupTimeout = time.Second * 10
	defaultMaxHeaderSize  = 1024 * 1024
	defaultHTTPPort       = 8080
	defaultHTTPSPort      = 8443
	defaultKeepAlive      = time.Minute * 3
	defaultReadTimeout    = time.Second * 30
	defaultWriteTimeout   = time.Second * 60
	defaultHost           = "127.0.0.1"
)

var defaultSchemes []string

func init() {
	defaultSchemes = []string{
		schemeHTTP,
	}
}

// NewServer creates a HTTP server with given handler as a default handler
func NewServer(handler http.Handler) *Server {
	s := new(Server)

	s.Name = "Dispatch HTTP server"
	s.shutdown = make(chan struct{})
	s.ready = make(chan struct{})
	s.handler = handler
	s.Logger = log.StandardLogger()

	s.Host = defaultHost
	s.CleanupTimeout = defaultCleanupTimeout
	s.MaxHeaderSize = defaultMaxHeaderSize

	// http default configuration
	s.Port = defaultHTTPPort
	s.KeepAlive = defaultKeepAlive
	s.ReadTimeout = defaultReadTimeout
	s.WriteTimeout = defaultWriteTimeout

	// https default configuration
	s.TLSPort = defaultHTTPSPort
	s.TLSKeepAlive = defaultKeepAlive
	s.TLSReadTimeout = defaultReadTimeout
	s.TLSWriteTimeout = defaultWriteTimeout

	return s
}

// Server provides an HTTP server with reasonable defaults, ability to handle both http and https, and proper shutdown.
type Server struct {
	// Name of the server.
	Name string
	// Logger configures the logger to use.
	Logger log.FieldLogger
	// EnabledListeners set the listeners to enable.
	EnabledListeners []string
	// CleanupTimeout is a grace period for which to wait before shutting down the server.
	CleanupTimeout time.Duration
	// MaxHeaderSize controls the maximum number of bytes the server will read parsing the request header's keys and values,
	// including the request line. It does not limit the size of the request body.
	MaxHeaderSize int
	// Host (or IP) to listen on.
	Host string

	// Port to listen on for plain HTTP connections.
	Port int
	// ListenLimit sets the maximum number of outstanding requests.
	ListenLimit int
	// KeepAlive sets the TCP keep-alive timeouts on accepted plain HTTP connections. It prunes dead TCP connections.
	KeepAlive time.Duration
	// ReadTimeout sets maximum duration before timing out read of the request for plain HTTP connections.
	ReadTimeout time.Duration
	// WriteTimeout sets the maximum duration before timing out write of the response for plain HTTP connections.
	WriteTimeout time.Duration
	httpListener net.Listener

	// TLSPort sets the port to listen on for HTTPS connections.
	TLSPort int
	// TLSCertificate sets the certificate file path to use for HTTPS connections.
	TLSCertificate string
	// TLSCertificateKey sets the private key file path to use for HTTPS connections.
	TLSCertificateKey string
	// TLSCACertificate sets the certificate authority file path to be used for HTTPS connections. Use only when verifying client certificate.
	TLSCACertificate string
	// TLSListenLimit set the maximum number of outstanding requests for HTTPS connections.
	TLSListenLimit int
	// TLSKeepAlive sets the TCP keep-alive timeouts on accepted HTTPS connections. It prunes dead TCP connections.
	TLSKeepAlive time.Duration
	// TLSReadTimeout sets maximum duration before timing out read of the request for HTTPS connections.
	TLSReadTimeout time.Duration
	// TLSWriteTimeout sets the maximum duration before timing out write of the response for HTTPS connections.
	TLSWriteTimeout time.Duration
	LetsEncrypt     bool
	Production      bool
	Domain          string
	httpsListener   net.Listener

	handler         http.Handler
	hasListeners    bool
	ready, shutdown chan struct{}
	shuttingDown    int32
}

func (s *Server) hasScheme(scheme string) bool {
	schemes := s.EnabledListeners
	if len(schemes) == 0 {
		schemes = defaultSchemes
	}

	for _, v := range schemes {
		if v == scheme {
			return true
		}
	}
	return false
}

// Serve the api
func (s *Server) Serve() (err error) {
	if !s.hasListeners {
		if err = s.Listen(); err != nil {
			return err
		}
	}

	// set default handler, if none is set
	if s.handler == nil {
		return errors.New("handler not set")
	}

	var wg sync.WaitGroup

	if s.hasScheme(schemeHTTP) {
		httpServer := new(http.Server)
		httpServer.MaxHeaderBytes = int(s.MaxHeaderSize)
		httpServer.ReadTimeout = s.ReadTimeout
		httpServer.WriteTimeout = s.WriteTimeout
		httpServer.SetKeepAlivesEnabled(int64(s.KeepAlive) > 0)
		if s.ListenLimit > 0 {
			s.httpListener = netutil.LimitListener(s.httpListener, s.ListenLimit)
		}

		if int64(s.CleanupTimeout) > 0 {
			httpServer.IdleTimeout = s.CleanupTimeout
		}

		httpServer.Handler = s.handler

		wg.Add(2)
		s.Logger.Infof("%s: serving HTTP traffic at http://%s", s.Name, s.httpListener.Addr())
		go func(l net.Listener) {
			defer wg.Done()
			if err := httpServer.Serve(l); err != nil && err != http.ErrServerClosed {
				s.Logger.Errorf("%v", err)
			}
			s.Logger.Infof("%s: stopped serving HTTP traffic at http://%s", s.Name, l.Addr())
		}(s.httpListener)
		go s.handleShutdown(&wg, httpServer)
	}

	if s.hasScheme(schemeHTTPS) {
		if s.LetsEncrypt {
			if s.Domain == "" {
				return errors.New("Need a domain name, use --domain <Domain Name>")
			}
			uDomain := strings.Replace(s.Domain, ".", "_", -1)
			if s.existCertificate(uDomain) {
				s.TLSCertificate = "./lets_encrypt/" + uDomain + ".crt"
				s.TLSCertificateKey = "./lets_encrypt/" + uDomain + ".key"
			} else {
				if err := s.getLetsEncrypt(uDomain); err != nil {
					return errors.Wrap(err, "Cannot get Let's Encrypt certificate")
				}
			}
		}

		if s.TLSCertificateKey == "" && s.TLSCertificate == "" {
			log.Warnf("HTTPS requested but key and cert paths are empty. using self-generated PKI")
			s.TLSCertificateKey, s.TLSCertificate, err = GeneratePKI([]string{s.Host})
			if err != nil {
				return errors.Wrap(err, "error generating key and certificate pair")
			}
		}
		if s.TLSCertificate == "" {
			return errors.New("missing TLS Certificate file")
		}
		if s.TLSCertificateKey == "" {
			return errors.New("missing TLS Certificate private key file")
		}

		httpsServer := new(http.Server)
		httpsServer.MaxHeaderBytes = int(s.MaxHeaderSize)
		httpsServer.ReadTimeout = s.TLSReadTimeout
		httpsServer.WriteTimeout = s.TLSWriteTimeout
		httpsServer.SetKeepAlivesEnabled(int64(s.TLSKeepAlive) > 0)
		if s.TLSListenLimit > 0 {
			s.httpsListener = netutil.LimitListener(s.httpsListener, s.TLSListenLimit)
		}
		if int64(s.CleanupTimeout) > 0 {
			httpsServer.IdleTimeout = s.CleanupTimeout
		}
		httpsServer.Handler = s.handler

		// Inspired by https://blog.bracebin.com/achieving-perfect-ssl-labs-score-with-go
		httpsServer.TLSConfig = &tls.Config{
			// Causes servers to use Go's default ciphersuite preferences,
			// which are tuned to avoid attacks. Does nothing on clients.
			PreferServerCipherSuites: true,
			// Only use curves which have assembly implementations
			// https://github.com/golang/go/tree/master/src/crypto/elliptic
			CurvePreferences: []tls.CurveID{tls.CurveP256},
			// Use modern tls mode https://wiki.mozilla.org/Security/Server_Side_TLS#Modern_compatibility
			NextProtos: []string{"http/1.1", "h2"},
			// https://www.owasp.org/index.php/Transport_Layer_Protection_Cheat_Sheet#Rule_-_Only_Support_Strong_Protocols
			MinVersion: tls.VersionTLS12,
			// These ciphersuites support Forward Secrecy: https://en.wikipedia.org/wiki/Forward_secrecy
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		}

		httpsServer.TLSConfig.Certificates = make([]tls.Certificate, 1)
		httpsServer.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair(string(s.TLSCertificate), string(s.TLSCertificateKey))

		if s.TLSCACertificate != "" {
			caCert, caCertErr := ioutil.ReadFile(string(s.TLSCACertificate))
			if caCertErr != nil {
				log.Fatal(caCertErr)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			httpsServer.TLSConfig.ClientCAs = caCertPool
			httpsServer.TLSConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}

		httpsServer.TLSConfig.BuildNameToCertificate()

		if err != nil {
			return err
		}

		wg.Add(2)
		s.Logger.Infof("%s: serving HTTPS traffic at https://%s", s.Name, s.httpsListener.Addr())
		go func(l net.Listener) {
			defer wg.Done()
			if err := httpsServer.Serve(l); err != nil && err != http.ErrServerClosed {
				s.Logger.Errorf("%v", err)
			}
			s.Logger.Infof("%s: stopped serving HTTPS traffic at https://%s", s.Name, l.Addr())
		}(tls.NewListener(s.httpsListener, httpsServer.TLSConfig))
		go s.handleShutdown(&wg, httpsServer)
	}
	// We finished initialization, report readiness
	close(s.ready)
	wg.Wait()
	return nil
}

// Listen creates the listeners for the server
func (s *Server) Listen() error {
	if s.hasListeners { // already done this
		return nil
	}
	if s.hasScheme(schemeHTTP) {
		listener, err := net.Listen("tcp", net.JoinHostPort(s.Host, strconv.Itoa(s.Port)))
		if err != nil {
			return err
		}

		h, p, err := swag.SplitHostPort(listener.Addr().String())
		if err != nil {
			return err
		}
		s.Host = h
		s.Port = p
		s.httpListener = listener
	}

	if s.hasScheme(schemeHTTPS) {
		tlsListener, err := net.Listen("tcp", net.JoinHostPort(s.Host, strconv.Itoa(s.TLSPort)))
		if err != nil {
			return err
		}

		sh, sp, err := swag.SplitHostPort(tlsListener.Addr().String())
		if err != nil {
			return err
		}
		s.Host = sh
		s.TLSPort = sp
		s.httpsListener = tlsListener
	}

	s.hasListeners = true
	return nil
}

// Wait waits until server is initialized
func (s *Server) Wait() {
	<-s.ready
}

// Shutdown server and clean up resources
func (s *Server) Shutdown() error {
	if atomic.LoadInt32(&s.shuttingDown) != 0 {
		return nil
	}
	atomic.AddInt32(&s.shuttingDown, 1)
	close(s.shutdown)
	return nil
}

func (s *Server) handleShutdown(wg *sync.WaitGroup, server *http.Server) {
	defer wg.Done()
	ctx, cancel := context.WithTimeout(context.TODO(), 15*time.Second)
	defer cancel()

	<-s.shutdown
	if err := server.Shutdown(ctx); err != nil {
		// Error from closing listeners, or context timeout:
		s.Logger.Errorf("%s: error when shutting down HTTP server: %v", s.Name, err)
	}
}

// GetHandler returns a handler useful for testing
func (s *Server) GetHandler() http.Handler {
	return s.handler
}

// SetHandler allows for setting a http handler on this server
func (s *Server) SetHandler(handler http.Handler) {
	s.handler = handler
}

// HTTPListener returns the http listener
func (s *Server) HTTPListener() (net.Listener, error) {
	if !s.hasListeners {
		if err := s.Listen(); err != nil {
			return nil, err
		}
	}
	return s.httpListener, nil
}

// TLSListener returns the https listener
func (s *Server) TLSListener() (net.Listener, error) {
	if !s.hasListeners {
		if err := s.Listen(); err != nil {
			return nil, err
		}
	}
	return s.httpsListener, nil
}

// HTTPURL returns the http url server is available at
func (s *Server) HTTPURL() string {
	if s.hasScheme("http") {
		return fmt.Sprintf("http://%s:%d", s.Host, s.Port)
	}
	return ""
}

// HTTPSURL returns https url server is available at
func (s *Server) HTTPSURL() string {
	if s.hasScheme("http") {
		return fmt.Sprintf("https://%s:%d", s.Host, s.TLSPort)
	}
	return ""
}

func (s *Server) existCertificate(uDomain string) bool {
	if _, err := os.Stat("./lets_encrypt/" + uDomain + ".key"); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat("./lets_encrypt/" + uDomain + ".crt"); os.IsNotExist(err) {
		return false
	}
	return true
}

func (s *Server) getLetsEncrypt(uDomain string) error {
	letsEncryptURL := utils.LetsEncryptStaging
	certDirectory := "./lets_encrypt"
	if s.Production {
		letsEncryptURL = utils.LetsEncryptProduction
	}
	err := os.MkdirAll(certDirectory, 0644)
	if err != nil {
		return err
	}
	// Generate account key
	aKey, err := newKey(path.Join(certDirectory, "account.pem"))
	if err != nil {
		return err
	}

	// Create the ACME client
	c := &acme.Client{
		Key:          aKey,
		DirectoryURL: letsEncryptURL,
	}

	// Begin with registration
	log.Infof("Attempting registration...")
	_, err = c.Register(context.TODO(), nil, func(string) bool { return true })
	if err != nil {
		return err
	}
	log.Infof("Registration succeeded")

	// Attempt authorization
	auth, err := c.Authorize(context.TODO(), s.Domain)
	if err != nil {
		return err
	}
	var challenge *acme.Challenge
	for _, c := range auth.Challenges {
		if c.Type == "http-01" {
			challenge = c
		}
	}
	if challenge == nil {
		return err
	}

	// Determine the correct path to listen on
	cPath := c.HTTP01ChallengePath(challenge.Token)
	cResponse, err := c.HTTP01ChallengeResponse(challenge.Token)
	if err != nil {
		return err
	}

	// Create a server that responds to the request
	mux := http.NewServeMux()
	mux.HandleFunc(cPath, func(w http.ResponseWriter, r *http.Request) {
		b := []byte(cResponse)
		w.Header().Set("Content-Length", strconv.Itoa(len(b)))
		w.WriteHeader(http.StatusOK)
		w.Write(b)
	})
	l, err := net.Listen("tcp", ":80")
	if err != nil {
		return err
	}
	defer l.Close()
	go func() {
		http.Serve(l, mux)
	}()

	// Perform the challenge
	log.Print("Performing challenge...")
	_, err = c.Accept(context.TODO(), challenge)
	if err != nil {
		return err
	}

	// Wait for authorization to complete
	_, err = c.WaitAuthorization(context.TODO(), auth.URI)
	if err != nil {
		return err
	}
	log.Infof("Challenge completed")

	// Generate a key for the domain
	dKey, err := newKey(fmt.Sprintf("%s.key", path.Join(certDirectory, uDomain)))
	if err != nil {
		return err
	}

	// Create the CSR (certificate signing request)
	csr, err := x509.CreateCertificateRequest(
		rand.Reader,
		&x509.CertificateRequest{
			Subject: pkix.Name{CommonName: s.Domain},
		},
		dKey,
	)
	if err != nil {
		return err
	}

	// Send the CSR and obtain the certificate
	log.Infof("Signing the certificate")
	ders, _, err := c.CreateCert(context.TODO(), csr, 90*24*time.Hour, true)
	if err != nil {
		return err
	}
	log.Infof("Certificate signed!")

	// Write the certificate bundle to disk
	w, err := os.Create(path.Join(
		certDirectory, fmt.Sprintf("%s.crt", uDomain),
	))
	if err != nil {
		return err
	}
	defer w.Close()
	for _, der := range ders {
		err := pem.Encode(w, &pem.Block{
			Type:  utils.CertType,
			Bytes: der,
		})
		if err != nil {
			return err
		}
	}

	log.Infof("Complete!")

	s.TLSCertificateKey = fmt.Sprintf("%s.key", path.Join(certDirectory, uDomain))
	s.TLSCertificate = path.Join(certDirectory, fmt.Sprintf("%s.crt", uDomain))
	return nil
}

func newKey(filename string) (crypto.Signer, error) {
	log.Printf("generating %d-bit RSA key...", utils.KeyLength)
	k, err := rsa.GenerateKey(rand.Reader, utils.KeyLength)
	if err != nil {
		return nil, err
	}
	b := pem.EncodeToMemory(&pem.Block{
		Type:  utils.KeyType,
		Bytes: x509.MarshalPKCS1PrivateKey(k),
	})
	if err = ioutil.WriteFile(filename, b, 0600); err != nil {
		log.Fatal(err)
	}
	log.Print("generated RSA key")
	return k, nil
}
