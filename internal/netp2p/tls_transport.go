package netp2p

import (
	"context"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"math/big"
	"net"
	"strings"
	"time"
)

func normalizeFingerprint(fp string) string {
	fp = strings.TrimSpace(strings.ToLower(fp))
	fp = strings.ReplaceAll(fp, ":", "")
	return fp
}

func generateTLSListener(bindAddr, tokenSeed string) (net.Listener, string, time.Time, error) {
	priv, err := deterministicECDSAKey(tokenSeed)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	windowStart := time.Now().UTC().Truncate(72 * time.Hour)
	serial := deterministicSerial(tokenSeed)
	tmpl := &x509.Certificate{
		SerialNumber:          serial,
		NotBefore:             windowStart.Add(-time.Hour),
		NotAfter:              windowStart.Add(72 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	cert := tls.Certificate{
		Certificate: [][]byte{der},
		PrivateKey:  priv,
	}
	sum := sha256.Sum256(der)
	fingerprint := hex.EncodeToString(sum[:])
	cfg := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cert},
	}
	ln, err := tls.Listen("tcp", bindAddr, cfg)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	return ln, fingerprint, tmpl.NotAfter, nil
}

func deterministicECDSAKey(seed string) (*ecdsa.PrivateKey, error) {
	if strings.TrimSpace(seed) == "" {
		return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	}
	curve := elliptic.P256()
	params := curve.Params()
	sum := sha256.Sum256([]byte(seed))
	n := new(big.Int).Sub(params.N, big.NewInt(1))
	d := new(big.Int).SetBytes(sum[:])
	d.Mod(d, n)
	d.Add(d, big.NewInt(1))
	// Pad d to 32 bytes (P-256 scalar size)
	dBytes := d.Bytes()
	if len(dBytes) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(dBytes):], dBytes)
		dBytes = padded
	}
	ecdhKey, err := ecdh.P256().NewPrivateKey(dBytes)
	if err != nil {
		return nil, errors.New("falha ao derivar chave TLS determinística")
	}
	// Convert crypto/ecdh key back to *ecdsa.PrivateKey
	ecdsaKey, err := ecdhToECDSA(ecdhKey)
	if err != nil {
		return nil, err
	}
	return ecdsaKey, nil
}

func ecdhToECDSA(key *ecdh.PrivateKey) (*ecdsa.PrivateKey, error) {
	rawPub := key.PublicKey().Bytes()
	// rawPub is uncompressed: 0x04 || X(32) || Y(32) = 65 bytes for P-256
	if len(rawPub) != 65 || rawPub[0] != 0x04 {
		return nil, errors.New("falha ao converter chave ECDH para ECDSA")
	}
	x := new(big.Int).SetBytes(rawPub[1:33])
	y := new(big.Int).SetBytes(rawPub[33:65])
	return &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     x,
			Y:     y,
		},
		D: new(big.Int).SetBytes(key.Bytes()),
	}, nil
}

func deterministicSerial(seed string) *big.Int {
	sum := sha256.Sum256([]byte("serial:" + seed))
	serial := new(big.Int).SetBytes(sum[:16])
	if serial.Sign() <= 0 {
		return big.NewInt(1)
	}
	return serial
}

func tlsClientConfig(inv InviteKey) (*tls.Config, error) {
	want := normalizeFingerprint(inv.Fingerprint)
	if want == "" {
		return nil, errors.New("fingerprint TLS ausente")
	}
	cfg := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true,
		VerifyPeerCertificate: func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
			if len(rawCerts) == 0 {
				return errors.New("certificado TLS ausente")
			}
			got := sha256.Sum256(rawCerts[0])
			if hex.EncodeToString(got[:]) != want {
				return errors.New("fingerprint TLS inválido")
			}
			return nil
		},
	}
	return cfg, nil
}

func dialSessionConn(inv InviteKey, timeout time.Duration) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: timeout}
	if strings.TrimSpace(inv.Fingerprint) == "" {
		return dialer.Dial("tcp", inv.Addr)
	}
	cfg, err := tlsClientConfig(inv)
	if err != nil {
		return nil, err
	}
	td := &tls.Dialer{
		NetDialer: dialer,
		Config:    cfg,
	}
	return td.DialContext(context.Background(), "tcp", inv.Addr)
}
