package grantprovider

import (
	"fmt"

	"github.com/spf13/cobra"
)

var requiredCommands = map[string]bool{
	"get-token": true,
	"get-url":   true,
}

type OAuth2Commands = map[string]*cobra.Command

func NewOAuth2Command(provider string, oauth2Commands OAuth2Commands) (*cobra.Command, error) {
	notFoundedCommands := []string{}
	availableCommands := []*cobra.Command{}
	for commandName := range requiredCommands {
		command, ok := oauth2Commands[commandName]
		if !ok {
			notFoundedCommands = append(notFoundedCommands, commandName)
		}
		availableCommands = append(availableCommands, command)
	}
	if len(notFoundedCommands) > 0 {
		return nil, fmt.Errorf("comandos requeridos no encontrados: %s", notFoundedCommands)
	}
	oauth2RootCommand := &cobra.Command{
		Use:   "oauth2",
		Short: fmt.Sprintf("Procesa operaciones de oauth2 para el proveedor %s", provider),
	}
	oauth2RootCommand.AddCommand(availableCommands...)
	return oauth2RootCommand, nil
}
