package grantprovider

import (
	"fmt"

	"github.com/spf13/cobra"
)

// requiredCommands define los comandos OAuth2 obligatorios que debe proporcionar cada proveedor.
var requiredCommands = map[string]bool{
	"get-token": true,
	"get-url":   true,
}

// OAuth2Commands es un mapa de comandos Cobra indexados por nombre.
// Debe incluir al menos los comandos definidos en requiredCommands.
type OAuth2Commands = map[string]*cobra.Command

// NewOAuth2Command crea un comando raíz "oauth2" para un proveedor específico.
// Verifica que oauth2Commands contenga todos los comandos requeridos (get-token, get-url).
// Retorna error si falta algún comando requerido.
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

// requiredGetURLParams define los parámetros obligatorios para get-url.
var requiredGetURLParams = []string{
	"response_type",
	"client_id",
	"redirect_uri",
	"scope",
	"state",
}

// requiredGetTokenParams define los parámetros obligatorios para get-token.
var requiredGetTokenParams = []string{
	"code",
}

// argumentsToMap convierte []CommandArgument a un mapa name->value para búsqueda eficiente.
func argumentsToMap(arguments []CommandArgument) map[string]string {
	result := make(map[string]string, len(arguments))
	for _, arg := range arguments {
		result[arg.Name] = arg.Value
	}
	return result
}

// findMissingParams devuelve la lista de parámetros de required que no están presentes en args.
func findMissingParams(required []string, args map[string]string) []string {
	var missing []string
	for _, param := range required {
		if _, ok := args[param]; !ok {
			missing = append(missing, param)
		}
	}
	return missing
}

// buildValidationError crea un ValidationError con violaciones para campos faltantes.
func buildValidationErrorForMissing(missing []string) ValidationError {
	violations := make([]FieldViolation, 0, len(missing))
	for _, field := range missing {
		violations = append(violations, FieldViolation{
			Field:     field,
			Namespace: "OAuth2Argument",
			Rule:      "required",
		})
	}
	return ValidationError{
		Violations: violations,
		cause:      fmt.Errorf("validation error: campos requeridos faltantes"),
	}
}

// ValidateOAuth2GetURL valida que los argumentos de get-url contengan los parámetros requeridos.
// Retorna ValidationError vacío si todo es válido, o ValidationError con Violations si faltan campos.
func ValidateOAuth2GetURL(arguments []CommandArgument) (ValidationError, error) {
	argsMap := argumentsToMap(arguments)
	missing := findMissingParams(requiredGetURLParams, argsMap)

	if len(missing) > 0 {
		return buildValidationErrorForMissing(missing), nil
	}

	return ValidationError{}, nil
}

// ValidateOAuth2GetToken valida que los argumentos de get-token contengan los parámetros requeridos.
// Retorna ValidationError vacío si todo es válido, o ValidationError con Violations si faltan campos.
func ValidateOAuth2GetToken(arguments []CommandArgument) (ValidationError, error) {
	argsMap := argumentsToMap(arguments)
	missing := findMissingParams(requiredGetTokenParams, argsMap)

	if len(missing) > 0 {
		return buildValidationErrorForMissing(missing), nil
	}

	return ValidationError{}, nil
}
