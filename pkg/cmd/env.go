package cmd

import (
	"os"
	"strings"

	"github.com/gookit/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewEnvCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "env",
		Short: "dumps environment variables and configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			data := [][]string{}
			for _, key := range viper.AllKeys() {
				envKey := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
				viperKey := "PERMIFY_" + envKey
				viperValue := obfuscateSecrets(viperKey, viper.GetString(key))
				envValue, _ := os.LookupEnv(viperKey)
				envValue = obfuscateSecrets(viperKey, envValue)
				data = append(data, []string{color.FgCyan.Render(viperKey), viperValue, envValue})
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Key", "Config Value", "Env Value"})
			table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_CENTER})
			table.SetCenterSeparator("|")
			for _, v := range data {
				table.Append(v)
			}
			table.Render()
			return nil
		},
	}
}

func obfuscateSecrets(key, value string) string {
	whitelist := []string{
		"PERMIFY_DATABASE_URI",
	}

	for _, wKey := range whitelist {
		if key == wKey {
			if lastColon := strings.LastIndex(value, ":"); lastColon != -1 {
				typeEnd := strings.Index(value[lastColon:], "/")
				if typeEnd != -1 {
					return value[:lastColon+typeEnd+1] + "***" + value[len(value)-3:]
				}
			}
			return value[:len(value)-3] + "***"
		}
	}
	return value
}
