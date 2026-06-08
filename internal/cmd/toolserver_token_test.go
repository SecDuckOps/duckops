package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUniqueNonEmptyTokens(t *testing.T) {
	t.Parallel()

	require.Equal(t, []string{"a", "b"}, uniqueNonEmptyTokens("a", " b ", "", "a", "b"))
	require.Equal(t, []string{"x", "y"}, prependUniqueToken("x", "y", "x"))
}
