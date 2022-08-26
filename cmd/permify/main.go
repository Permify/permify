package main

import (
	`fmt`
	"log"

	"github.com/Permify/permify/internal/app"
	"github.com/Permify/permify/internal/config"
)

const (
	version = "v0.0.0-alpha2"
	color   = "\033[0;37m%s\033[0m"
	banner  = `
██████╗ ███████╗██████╗ ███╗   ███╗██╗███████╗██╗   ██╗
██╔══██╗██╔════╝██╔══██╗████╗ ████║██║██╔════╝╚██╗ ██╔╝
██████╔╝█████╗  ██████╔╝██╔████╔██║██║█████╗   ╚████╔╝ 
██╔═══╝ ██╔══╝  ██╔══██╗██║╚██╔╝██║██║██╔══╝    ╚██╔╝  
██║     ███████╗██║  ██║██║ ╚═╝ ██║██║██║        ██║   
╚═╝     ╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝╚═╝╚═╝        ╚═╝   
High performance authorization service %s
`
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	log.Printf(color, fmt.Sprintf(banner, version))

	// Run
	app.Run(cfg)
}
