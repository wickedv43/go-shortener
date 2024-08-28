package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_main(t *testing.T) {
	assert.NotEmpty(t, S.Data)
}
