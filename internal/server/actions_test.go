package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShorting(t *testing.T) {
	require.NotEmpty(t, Shorting())
	require.Len(t, Shorting(), 8)
}
