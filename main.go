package main

import (
	"echo/routes"

	"github.com/google/gops/agent"
	"github.com/rs/zerolog/log"
)

// @title AUTHN API
// @version 2.0
// @description APIs for user authentication

// @contact.name API Support
// @contact.email hari@dmartlabs.com
// @host api.dmartlink.com
// @BasePath /auth
func main() {

	log.Info().Msg("Starting Echo service")

	if err := agent.Listen(agent.Options{Addr: ":8081"}); err != nil {
		log.Error().AnErr("Error", err)
	}

	routes.Start()
}
