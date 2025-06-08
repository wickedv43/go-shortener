package url

import (
	"testing"
)

func Benchmark_ShortURL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ShortURL()
	}
}
