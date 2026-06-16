package serverclient

type Token struct {
	UUID      string   `json:"uuid"`
	TokenName string   `json:"token_name"`
	Scopes    []string `json:"scopes"`
	ExpiresAt string   `json:"expires_at"`
	CreatedAt string   `json:"created_at"`
	LastUsed  string   `json:"last_used_at,omitempty"`
}

type CreateTokenPayload struct {
	TokenName     string   `json:"token_name"`
	Scopes        []string `json:"scopes"`
	ExpiresInDays int      `json:"expiresInDays"`
}

type TokensData struct {
	Token    *Token   `json:"token,omitempty"`
	Tokens   []Token  `json:"tokens,omitempty"`
	RawToken string   `json:"rawToken,omitempty"`
}

func (c *Client) ListTokens() ([]Token, error) {
	var data TokensData
	err := c.Get("/api/v1/tokens", &data)
	return data.Tokens, err
}

func (c *Client) GetToken(uuid string) (*Token, error) {
	var data TokensData
	err := c.Get("/api/v1/tokens/"+uuid, &data)
	return data.Token, err
}

func (c *Client) CreateToken(payload CreateTokenPayload) (*Token, string, error) {
	var data TokensData
	err := c.Post("/api/v1/tokens", payload, &data)
	return data.Token, data.RawToken, err
}

func (c *Client) DeleteToken(uuid string) error {
	return c.Delete("/api/v1/tokens/"+uuid, nil)
}
