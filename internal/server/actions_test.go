package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShortURL_BasicProperties(t *testing.T) {
	const count = 1000
	generated := make(map[string]bool)

	for i := 0; i < count; i++ {
		s := ShortURL()

		require.Len(t, s, 8, "short URL should be 8 chars long")

		for _, r := range s {
			require.Contains(t, letterBytes, string(r), "invalid character: %q", r)
		}

		// Простая проверка на коллизии
		if generated[s] {
			t.Logf("duplicate found: %s", s)
		}
		generated[s] = true
	}

	// Коллизий должно быть очень мало или вовсе не быть
	require.Greater(t, len(generated), 990, "too many duplicates in random generation")
}
