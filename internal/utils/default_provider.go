package utils

func DefaultProvider() string {
	cfg, err := LoadConfig()
	if err != nil || cfg.PrivateEndpoint == "" {
		return "adoptium"
	}
	return "private"
}
