package url

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestURLService_Gen(t *testing.T) {
	u := &URLService{}
	input := []string{"a", "b", "c"}
	ch := u.Gen(input...)

	var result []string
	for val := range ch {
		result = append(result, val)
	}

	require.ElementsMatch(t, input, result)
}

func TestURLService_FanIn(t *testing.T) {
	u := &URLService{}

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

	out := u.FanIn(ch1, ch2)

	var result []string
	for val := range out {
		result = append(result, val)
	}

	require.ElementsMatch(t, []string{"a", "b", "c", "d"}, result)
}
