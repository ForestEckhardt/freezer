package github

type Config struct {
	Endpoint string
	Token    string
}

func NewConfig(endpoint, token string) Config {
	return Config{
		Endpoint: endpoint,
		Token:    token,
	}
}
