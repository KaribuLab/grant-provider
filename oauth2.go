package grantprovider

import (
	"bytes"
	"fmt"
	"io"

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

// OAuth2CommandHandler extiende CommandHandler para flujos OAuth2.
// Los providers deben implementar esta interfaz: además de Invoke, deben exponer
// GetCredentialsService y SetCredentialsService para que OAuth2CommandInvoker
// inyecte el GetClientCredentialsService configurado con el OTT y el endpoint
// de exchange recibidos en cada InvokeCommand.
type OAuth2CommandHandler interface {
	CommandHandler
	// GetCredentialsService retorna el GetClientCredentialsService inyectado por OAuth2CommandInvoker.
	GetCredentialsService() GetClientCredentialsService
	// SetCredentialsService recibe el GetClientCredentialsService construido por ExchangeFetcherFactory.
	// Es llamado automáticamente por OAuth2CommandInvoker antes de invocar Invoke.
	SetCredentialsService(GetClientCredentialsService)
}

// OAuth2CommandInvoker extiende CommandInvoker para flujos OAuth2.
// Su Run decodifica el InvokeCommand del stdin, usa ExchangeFetcherFactory para
// construir el GetClientCredentialsService apropiado e inyectarlo en el handler
// vía SetCredentialsService, dejándolo listo para obtener credenciales cuando
// su método Invoke sea llamado.
type OAuth2CommandInvoker struct {
	CommandInvoker
	// ExchangeFetcherFactory construye el GetClientCredentialsService adecuado para
	// cada InvokeCommand, configurado con el provider, sessionID y exchangeEndpoint del comando.
	ExchangeFetcherFactory ExchangeFetcherFactory
}

// NewOAuth2CommandInvoker crea un OAuth2CommandInvoker garantizando en tiempo de
// construcción que el handler implementa OAuth2CommandHandler.
// Usar esta función en lugar de inicializar la estructura directamente evita el
// pánico que produciría un handler incompatible en tiempo de ejecución.
func NewOAuth2CommandInvoker(handler OAuth2CommandHandler, factory ExchangeFetcherFactory) *OAuth2CommandInvoker {
	return &OAuth2CommandInvoker{
		CommandInvoker:         *NewCommandInvoker(handler),
		ExchangeFetcherFactory: factory,
	}
}

// Run decodifica el InvokeCommand desde stdin, construye el GetClientCredentialsService
// con ExchangeFetcherFactory, lo inyecta en el handler vía SetCredentialsService y
// delega el resto del flujo (validación e Invoke) en CommandInvoker.Run.
// Retorna error si el handler no implementa OAuth2CommandHandler, el JSON es
// inválido o la serialización interna falla.
func (ci *OAuth2CommandInvoker) Run(stdin io.Reader) (InvokeResponse, error) {
	command, err := FromJSON[InvokeCommand](stdin)
	if err != nil {
		return InvokeResponse{}, err
	}

	handler, ok := ci.CommandInvoker.handler.(OAuth2CommandHandler)
	if !ok {
		return InvokeResponse{}, fmt.Errorf("el handler no implementa OAuth2CommandHandler")
	}
	handler.SetCredentialsService(ci.ExchangeFetcherFactory(command))

	// Re-serializar el comando para que CommandInvoker.Run pueda leerlo desde un io.Reader.
	writer := new(bytes.Buffer)
	if err = ToJSON(writer, command); err != nil {
		return InvokeResponse{}, err
	}
	return ci.CommandInvoker.Run(writer)
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
