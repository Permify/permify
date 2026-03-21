package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadCredentials_MissingFile(t *testing.T) {
	t.Parallel()
	_, err := LoadCredentials(filepath.Join(t.TempDir(), "nope"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "credentials file not found")
	assert.True(t, errors.Is(err, os.ErrNotExist))
}

func TestLoadCredentials_Valid(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")
	err := os.WriteFile(path, []byte("endpoint: localhost:3478\n"), 0o600)
	require.NoError(t, err)

	c, err := LoadCredentials(path)
	require.NoError(t, err)
	assert.Equal(t, "localhost:3478", c.Endpoint)
}

func TestLoadCredentials_RelativeTLSCAPath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	credPath := filepath.Join(dir, "credentials")
	err := os.WriteFile(credPath, []byte("endpoint: localhost:3478\ntls_ca_path: myca.pem\n"), 0o600)
	require.NoError(t, err)

	c, err := LoadCredentials(credPath)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "myca.pem"), c.TLSCAPath)
}

func TestLoadCredentials_MissingEndpoint(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")
	err := os.WriteFile(path, []byte("api_token: x\n"), 0o600)
	require.NoError(t, err)

	_, err = LoadCredentials(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint is required")
}

func TestGRPCDialOptions_InsecureWithoutToken(t *testing.T) {
	t.Parallel()
	opts, err := GRPCDialOptions(&CredentialsFile{Endpoint: "x", APIToken: ""})
	require.NoError(t, err)
	assert.Len(t, opts, 1)
}

func TestGRPCDialOptions_TLSWithToken(t *testing.T) {
	t.Parallel()
	opts, err := GRPCDialOptions(&CredentialsFile{Endpoint: "x", APIToken: "secret"})
	require.NoError(t, err)
	assert.Len(t, opts, 2)
}

func TestGRPCDialOptions_InvalidCA(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	caPath := filepath.Join(dir, "ca.pem")
	err := os.WriteFile(caPath, []byte("not pem"), 0o600)
	require.NoError(t, err)

	_, err = GRPCDialOptions(&CredentialsFile{APIToken: "t", TLSCAPath: caPath})
	require.Error(t, err)
}
