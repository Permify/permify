package cmd

import (
	"os"
	"strings"

	"github.com/Permify/permify/pkg/cmd/flags"

	"github.com/gookit/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect permify configuration and environment variables ",
		RunE: func(cmd *cobra.Command, args []string) error {
			data := prepareConfigData(cmd)
			renderConfigTable(data)
			return nil
		},
	}

	flags.RegisterServeFlags(cmd)

	return cmd
}

func prepareConfigData(cmd *cobra.Command) [][]string {
	data := [][]string{}
	for _, key := range viper.AllKeys() {
		viperKey, viperValue, source := getConfigDetails(key, cmd)
		data = append(data, []string{color.FgCyan.Render(viperKey), viperValue, source})
	}
	return data
}

func getConfigDetails(key string, cmd *cobra.Command) (string, string, string) {
	envKey := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
	viperKey := "PERMIFY_" + envKey
	viperValue := viper.GetString(key)
	envValue, envExists := os.LookupEnv(viperKey)

	source := getKeyOrigin(envExists, cmd, key)
	value := viperValue
	if envExists {
		value = envValue
	}
	return viperKey, obfuscateSecrets(viperKey, value), source
}

func getKeyOrigin(envExists bool, cmd *cobra.Command, key string) string {
	if cmd.Flags().Changed(strings.ReplaceAll(key, ".", "-")) {
		return color.FgLightGreen.Render("flag")
	}
	if envExists {
		return color.FgLightBlue.Render("env")
	}
	return color.FgYellow.Render("file")
}

func renderConfigTable(data [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Key", "Value", "Source"})
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER})
	table.SetCenterSeparator("|")
	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}

func obfuscateSecrets(key, value string) string {
	secrets := []string{
		"PERMIFY_DATABASE_URI",
	}

	for _, wKey := range secrets {
		if key == wKey {
			if len(value) < 3 {
				return value
			}
			if lastColon := strings.LastIndex(value, ":"); lastColon != -1 {
				typeEnd := strings.Index(value[lastColon:], "/")
				if typeEnd != -1 {
					if len(value)-3 > lastColon+typeEnd+1 {
						return value[:lastColon+typeEnd+1] + "***" + value[len(value)-3:]
					}
					return value
				}
			}
			return value[:len(value)-3] + "***"
		}
	}
	return value
}
