package models

// ModuleConfig 模块配置结构
type ModuleConfig struct {
	Interface struct {
		PrivateKey string `json:"private_key"`
		Address    string `json:"address"`
		DNS        string `json:"dns,omitempty"`
	} `json:"interface"`

	Peer struct {
		PublicKey           string   `json:"public_key"`
		Endpoint            string   `json:"endpoint"`
		AllowedIPs          []string `json:"allowed_ips"`
		PersistentKeepalive int      `json:"persistent_keepalive"`
	} `json:"peer"`
}

// PeerConfig 运维端配置结构
type PeerConfig struct {
	PublicKey  string   `json:"public_key"`
	AllowedIPs []string `json:"allowed_ips"`
	Comment    string   `json:"comment,omitempty"`
}
