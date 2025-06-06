package logger

import (
	"testing"

	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestNewLogger_Success(t *testing.T) {
	container := do.New()

	// Биндим пустой Logger для успешного вызова do.InvokeStruct
	do.Provide(container, func(i do.Injector) (*Logger, error) {
		return &Logger{}, nil
	})

	l, err := NewLogger(container)
	require.NoError(t, err)
	require.NotNil(t, l)
	require.Equal(t, logrus.InfoLevel, l.Logger.GetLevel())
}

func TestNewLogger_InvokeError(t *testing.T) {
	container := do.New()

	l, err := NewLogger(container)
	require.NotNil(t, l)
	require.NoError(t, err)
}
