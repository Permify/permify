package main

import (
	"fmt"
	"log"

	"github.com/Permify/permify/internal/app"
	"github.com/Permify/permify/internal/config"
)

const (
	// Version of Permify
	Version = "v0.0.0-alpha6"
	color   = "\033[0;37m%s\033[0m"
	banner  = `

██████╗ ███████╗██████╗ ███╗   ███╗██╗███████╗██╗   ██╗
██╔══██╗██╔════╝██╔══██╗████╗ ████║██║██╔════╝╚██╗ ██╔╝
██████╔╝█████╗  ██████╔╝██╔████╔██║██║█████╗   ╚████╔╝ 
██╔═══╝ ██╔══╝  ██╔══██╗██║╚██╔╝██║██║██╔══╝    ╚██╔╝  
██║     ███████╗██║  ██║██║ ╚═╝ ██║██║██║        ██║   
╚═╝     ╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝╚═╝╚═╝        ╚═╝   
_______________________________________________________
Fine-grained Authorization System %s
`
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	log.Printf(color, fmt.Sprintf(banner, Version))

	// Run
	app.Run(cfg)
}
