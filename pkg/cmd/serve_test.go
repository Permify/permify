package cmd

import (
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/Permify/permify/internal/config"
	"github.com/Permify/permify/pkg/cmd/flags"
)

func TestShutdownDelayFlag(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	cmd := NewServeCommand()
	assert.NoError(t, cmd.ParseFlags([]string{"--server-shutdown-delay=5s"}))
	flags.RegisterServeFlags(cmd.Flags())

	var cfg config.Config
	assert.NoError(t, viper.Unmarshal(&cfg))
	assert.Equal(t, 5*time.Second, cfg.Server.ShutdownDelay)
}

func TestShutdownDelayEnvVar(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	t.Setenv("PERMIFY_SHUTDOWN_DELAY", "10s")

	cmd := NewServeCommand()
	flags.RegisterServeFlags(cmd.Flags())

	var cfg config.Config
	assert.NoError(t, viper.Unmarshal(&cfg))
	assert.Equal(t, 10*time.Second, cfg.Server.ShutdownDelay)
}
