package server

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestShorting(t *testing.T) {
	require.NotEmpty(t, Shorting())
	require.Len(t, Shorting(), 8)
}
