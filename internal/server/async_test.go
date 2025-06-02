package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServer_gen(t *testing.T) {
	s := &Server{}
	input := []string{"a", "b", "c"}
	ch := s.gen(input...)

	var result []string
	for val := range ch {
		result = append(result, val)
	}

	require.ElementsMatch(t, input, result)
}

func TestServer_fanIn(t *testing.T) {
	s := &Server{}

	ch1 := make(chan string)
	ch2 := make(chan string)

	go func() {
		defer close(ch1)
		ch1 <- "a"
		ch1 <- "b"
	}()

	go func() {
		defer close(ch2)
		ch2 <- "c"
		ch2 <- "d"
	}()

	out := s.fanIn(ch1, ch2)

	var result []string
	for val := range out {
		result = append(result, val)
	}

	require.ElementsMatch(t, []string{"a", "b", "c", "d"}, result)
}
