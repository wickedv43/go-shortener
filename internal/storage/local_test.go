package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocalStorage_SaveAndGet(t *testing.T) {
	st := &LocalStorage{}

	d := Data{UUID: 42, OriginalURL: "http://example.com", ShortURL: "abc"}
	err := st.Save(context.Background(), d)
	require.NoError(t, err)

	got, err := st.Get(context.Background(), "abc")
	require.NoError(t, err)
	require.Equal(t, d.OriginalURL, got.OriginalURL)

	got2, err := st.Get(context.Background(), "http://example.com")
	require.NoError(t, err)
	require.Equal(t, d.ShortURL, got2.ShortURL)
}

func TestLocalStorage_Get_NotFound(t *testing.T) {
	st := &LocalStorage{}

	_, err := st.Get(context.Background(), "notfound")
	require.ErrorContains(t, err, "not found")
}

func TestLocalStorage_GetAll(t *testing.T) {
	st := &LocalStorage{}
	_ = st.Save(context.Background(), Data{UUID: 1, ShortURL: "s1", OriginalURL: "http://1"})
	_ = st.Save(context.Background(), Data{UUID: 2, ShortURL: "s2", OriginalURL: "http://2"})
	_ = st.Save(context.Background(), Data{UUID: 1, ShortURL: "s3", OriginalURL: "http://3"})

	list, err := st.GetAll(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, list, 2)

	empty, err := st.GetAll(context.Background(), 999)
	require.ErrorContains(t, err, "no content")
	require.Len(t, empty, 0)
}

func TestLocalStorage_HealthCheck(t *testing.T) {
	st := &LocalStorage{}
	require.NoError(t, st.HealthCheck())
}

func TestLocalStorage_BatchDelete(t *testing.T) {
	st := &LocalStorage{}
	require.NoError(t, st.BatchDelete([]string{"abc"}))
}

func TestLocalStorage_Close(t *testing.T) {
	st := &LocalStorage{}
	require.NoError(t, st.Close())
}
