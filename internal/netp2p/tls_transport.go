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
	"net/url"
	"strings"
	"time"

	"truco-tui/internal/netrelay"
)

func normalizeFingerprint(fp string) string {
	fp = strings.TrimSpace(strings.ToLower(fp))
	fp = strings.ReplaceAll(fp, ":", "")
	return fp
}

func buildTLSConfig(tlsSeed string) (*tls.Config, string, time.Time, error) {
	priv, err := deterministicECDSAKey(tlsSeed)
	if err != nil {
		return nil, "", time.Time{}, err
	}
	windowStart := time.Now().UTC().Truncate(72 * time.Hour)
	serial := deterministicSerial(tlsSeed)
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
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{cert},
	}
	return cfg, fingerprint, tmpl.NotAfter, nil
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

func randomTLSSeed() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func tlsClientConfig(inv InviteKey) (*tls.Config, error) {
	want := normalizeFingerprint(inv.Fingerprint)
	if want == "" {
		return nil, errors.New("fingerprint TLS ausente")
	}
	cfg := &tls.Config{
		MinVersion:         tls.VersionTLS13,
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
	return dialSessionConnWithRelay(inv, timeout, "", "", "")
}

func dialSessionConnWithRelay(inv InviteKey, timeout time.Duration, playerName, desiredRole, playerSession string) (net.Conn, error) {
	transport := strings.TrimSpace(inv.Transport)
	if transport == "" {
		transport = "tcp_tls"
	}
	if transport == "relay_quic_v2" {
		sec := relaySecurityFromInvite(inv)
		joinResp, err := netrelay.JoinSession(inv.RelayURL, sec, netrelay.JoinSessionRequest{
			SessionID:     inv.RelaySessionID,
			JoinTicket:    inv.RelayJoinTicket,
			PlayerName:    playerName,
			DesiredRole:   desiredRole,
			PlayerSession: playerSession,
		})
		if err != nil {
			return nil, err
		}
		target := strings.TrimSpace(joinResp.AuthorityPeerID)
		if target == "" {
			target = strings.TrimSpace(inv.RelayAuthorityPeer)
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		raw, err := netrelay.OpenPeerTunnel(ctx, sec, joinResp.QuicAddr, inv.RelaySessionID, joinResp.PeerID, joinResp.PeerCredential, target)
		if err != nil {
			return nil, err
		}
		return wrapTLSClient(raw, inv)
	}

	dialer := &net.Dialer{Timeout: timeout}
	raw, err := dialer.Dial("tcp", inv.Addr)
	if err != nil {
		return nil, err
	}
	return wrapTLSClient(raw, inv)
}

func relaySecurityFromInvite(inv InviteKey) netrelay.ClientSecurity {
	sec := netrelay.ClientSecurity{
		RelaySPKIPin: strings.TrimSpace(inv.RelaySPKIPin),
	}
	if u, err := url.Parse(strings.TrimSpace(inv.RelayURL)); err == nil {
		sec.ServerName = strings.TrimSpace(u.Hostname())
	}
	return sec
}

func wrapTLSClient(raw net.Conn, inv InviteKey) (net.Conn, error) {
	cfg, err := tlsClientConfig(inv)
	if err != nil {
		closeConnWithLog(raw, "wrap tls config failure")
		return nil, err
	}
	tconn := tls.Client(raw, cfg)
	if err := tconn.Handshake(); err != nil {
		closeConnWithLog(raw, "tls client handshake")
		return nil, err
	}
	return tconn, nil
}
