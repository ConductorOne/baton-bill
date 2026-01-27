package main

import (
	cfg "github.com/ConductorOne/baton-bill/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/config"
)

func main() {
	config.Generate("bill", cfg.Config)
}
