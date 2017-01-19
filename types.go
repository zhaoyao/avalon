package main

// Backend holds backend configuration.
type Backend struct {
	Servers      []Server      `json:"servers,omitempty"`
	LoadBalancer *LoadBalancer `json:"loadBalancer,omitempty"`
	MaxConn      int           `json:"maxConn,omitempty"`
}

// Server holds server configuration.
type Server struct {
	URL    string `json:"url,omitempty"` // eg: tcp://192.168.1.100:4321 or http://192.168.1.100:4322/hprose
	Weight int    `json:"weight"`
}

// LoadBalancer holds load balancing configuration.
type LoadBalancer struct {
	Method string `json:"method,omitempty"`
	Sticky bool   `json:"sticky,omitempty"`
}

type Configuration struct {
	Backends map[string]*Backend `json:"backends"`
}
