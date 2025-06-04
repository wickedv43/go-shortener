package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetGitCommit(t *testing.T) {
	commit := getGitCommit()

	require.NotEmpty(t, commit)
	require.LessOrEqual(t, len(commit), 12) // короткий SHA обычно до 12 символов
	require.Regexp(t, "^[a-f0-9]+$", commit)
}
