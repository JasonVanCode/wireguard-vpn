package models

// WireGuardKey WireGuard密钥对
type WireGuardKey struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}
