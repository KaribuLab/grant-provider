package grantprovider

import (
	"bytes"
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

// OAuth2CommandHandler extiende CommandHandler con la capacidad de construir el ExchangeFetcher
// necesario para obtener credenciales del cliente durante la ejecución de un comando OAuth2.
// Los providers deben implementar esta interfaz para integrarse con GetClientCredentialsService.
type OAuth2CommandHandler interface {
	CommandHandler
	// GetExecutorFetcher construye el ExchangeFetcher que será usado por GetClientCredentialsService
	// para intercambiar el OTT por credenciales en el endpoint de exchange.
	GetExecutorFetcher(provider string, sessionID string, exchangeEndpoint string) ExchangeFetcher
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

// ClientCredentialsData contiene las credenciales del cliente OAuth2 retornadas por el exchange.
type ClientCredentialsData struct {
	ClientID     string `json:"client_id" validate:"required"`
	ClientSecret string `json:"client_secret" validate:"required"`
}

// GetClientCredentialsService obtiene ClientCredentialsData a través de un ExchangeFetcher.
// Al depender de la interfaz ExchangeFetcher, puede probarse sin servidor HTTP usando un mock.
type GetClientCredentialsService struct {
	// ExchangeFetcher es la implementación que realiza el intercambio efectivo del OTT.
	// En producción usar ExchangeFetcherService; en tests usar un mock.
	ExchangeFetcher
}

// Execute intercambia el OTT por credenciales de cliente.
// Llama a ExchangeFetcher.Execute y decodifica ExchangeReponse.Data en ClientCredentialsData.
func (g *GetClientCredentialsService) Execute(exchangeRequest ExchangeRequest) (ClientCredentialsData, error) {
	exchangeResponse, err := g.ExchangeFetcher.Execute(exchangeRequest)
	if err != nil {
		return ClientCredentialsData{}, err
	}
	data := new(bytes.Buffer)
	err = ToJSON(data, exchangeResponse.Data)
	if err != nil {
		return ClientCredentialsData{}, err
	}
	clientCredentials, err := FromJSON[ClientCredentialsData](data)
	if err != nil {
		return ClientCredentialsData{}, err
	}
	return clientCredentials, nil
}
