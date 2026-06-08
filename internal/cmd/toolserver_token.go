package cmd

import (
	"log/slog"
	"os"
	"strings"
)

func readCachedToolServerToken() (string, error) {
	data, err := os.ReadFile(toolServerTokenPath())
	if err != nil {
		return "", err
	}
	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", os.ErrNotExist
	}
	return token, nil
}

func cacheToolServerToken(token string) {
	token = strings.TrimSpace(token)
	if token == "" {
		return
	}
	if err := os.MkdirAll(duckopsGlobalDir(), 0o700); err != nil {
		slog.Error("Failed to create ~/.duckops directory", "error", err)
		return
	}
	if err := writeToolServerToken(token); err != nil {
		slog.Error("Failed to cache tool server token", "error", err)
	}
}

// resolveVerifiedToolServerToken returns the first candidate that authenticates
// against the running tool server. When a Docker sandbox is available, the
// container token is checked first because it is authoritative.
func resolveVerifiedToolServerToken(port int, candidates ...string) (string, bool) {
	ordered := uniqueNonEmptyTokens(candidates...)

	if os.Getenv("DUCKOPS_SANDBOX_MODE") != "true" {
		if recovered := recoverTokenFromContainer(containerName); recovered != "" {
			ordered = prependUniqueToken(recovered, ordered...)
		}
	}

	for _, token := range ordered {
		switch verifyTokenResult(port, token) {
		case "valid":
			cacheToolServerToken(token)
			return token, true
		}
	}

	// Old tool servers may not expose /verify. Trust the container token in that
	// case because it is the only source of truth for docker-managed sandboxes.
	if os.Getenv("DUCKOPS_SANDBOX_MODE") != "true" {
		if recovered := recoverTokenFromContainer(containerName); recovered != "" {
			if verifyTokenResult(port, recovered) == "unknown" {
				cacheToolServerToken(recovered)
				return recovered, true
			}
		}
	}

	return "", false
}

func uniqueNonEmptyTokens(tokens ...string) []string {
	seen := make(map[string]struct{}, len(tokens))
	out := make([]string, 0, len(tokens))
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		out = append(out, token)
	}
	return out
}

func prependUniqueToken(token string, rest ...string) []string {
	return uniqueNonEmptyTokens(append([]string{token}, rest...)...)
}
