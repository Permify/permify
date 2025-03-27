package internal

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// Identifier is the unique identifier for the Permify.
var Identifier = ""

/*
✨ OneLiner: Open-source authorization service inspired by Google Zanzibar

📚 Docs: https://docs.permify.co
🐙 GitHub: https://github.com/Permify/permify
📝 Blog: https://permify.co/blog

💬 Discord: https://discord.gg/n6KfzYxhPp
🐦 Twitter: https://twitter.com/GetPermify
💼 LinkedIn: https://www.linkedin.com/company/permifyco
*/
const (
	// Version is the last release of the Permify (e.g. v0.1.0)
	Version = "v1.3.6"
)

// Function to create a single line of the ASCII art with centered content and color
func createLine(content string, totalWidth int, borderColor, contentColor *color.Color) string {
	contentLength := len(content)
	paddingWidth := (totalWidth - contentLength - 4) / 2
	if paddingWidth < 0 {
		paddingWidth = 0
	}
	leftPadding := strings.Repeat(" ", paddingWidth)
	rightPadding := strings.Repeat(" ", totalWidth-2-contentLength-paddingWidth)
	border := borderColor.Sprint("│")
	contentWithColor := contentColor.Sprintf("%s%s%s", leftPadding, content, rightPadding)
	return fmt.Sprintf("%s%s%s", border, contentWithColor, border)
}

func PrintBanner() {
	borderColor := color.New(color.FgWhite)
	contentColor := color.New(color.FgWhite)

	versionInfo := fmt.Sprintf("Permify %s", Version)

	lines := []string{
		borderColor.Sprint("┌────────────────────────────────────────────────────────┐"),
		createLine(versionInfo, 58, borderColor, color.New(color.FgBlue)),
		createLine("Fine-grained Authorization Service", 58, borderColor, contentColor),
		createLine("", 58, borderColor, contentColor),
		createLine("docs: ............... https://docs.permify.co", 58, borderColor, contentColor),
		createLine("github: .. https://github.com/Permify/permify", 58, borderColor, contentColor),
		createLine("blog: ............... https://permify.co/blog", 58, borderColor, contentColor),
		createLine("", 58, borderColor, contentColor),
		borderColor.Sprint("└────────────────────────────────────────────────────────┘"),
	}

	for _, line := range lines {
		fmt.Println(line)
	}
}
