package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"testing"
	"time"

	"chat-backend/internal/model"
	"chat-backend/internal/testutil"
	"chat-backend/internal/ws"
)

func TestServer_HTTPS(t *testing.T) {
	store := &testutil.MockStore{
		LastFn: func(_ context.Context, n int) ([]model.Message, error) {
			return []model.Message{{Username: "eve", Content: "hey"}}, nil
		},
	}
	ps := &testutil.MockPubSub{}
	hub := ws.NewHub(4)
	handler := ws.NewHandler(hub, store, ps, "*")

	mux := http.NewServeMux()
	mux.HandleFunc("/messages", handler.History)

	srv := &http.Server{
		Addr:    "127.0.0.1:0",
		Handler: mux,
	}

	certPair, err := generateSelfSigned()
	if err != nil {
		t.Fatalf("cert gen: %v", err)
	}
	srv.TLSConfig = &tls.Config{Certificates: []tls.Certificate{certPair}}

	ln, err := tls.Listen("tcp", srv.Addr, srv.TLSConfig)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	go srv.Serve(ln)
	defer srv.Shutdown(context.Background())

	cl := &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		Timeout:   5 * time.Second,
	}
	resp, err := cl.Get("https://" + ln.Addr().String() + "/messages")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()

	var msgs []model.Message
	_ = json.NewDecoder(resp.Body).Decode(&msgs)
	if len(msgs) != 1 || msgs[0].Username != "eve" {
		t.Fatalf("unexpected body: %+v", msgs)
	}
}

func generateSelfSigned() (tls.Certificate, error) {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "chat-test",
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().Add(24 * time.Hour),

		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	template.DNSNames = []string{"localhost"}
	template.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyBytes, _ := x509.MarshalECPrivateKey(priv)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

	return tls.X509KeyPair(certPEM, keyPEM)
}
