# grant-provider

Librería en Go para **invocar comandos** con contrato JSON, **validación** basada en etiquetas `validate` y utilidades de **configuración** bajo el directorio del usuario (`~/.grant`).

## Requisitos

- Go **1.25** o compatible con el `go` de tu `go.mod`.

## Instalación

```bash
go get github.com/KaribuLab/grant-provider
```

## Conceptos principales

| Pieza | Rol |
|--------|-----|
| [`InvokeCommand`](invoke.go) | Entrada: comando, proveedor, sesión, OTT, endpoint de exchange y argumentos opcionales. |
| [`InvokeResponse`](invoke.go) | Salida: `result` embebido, `data` opcional (`any`) y `additional_data` opcional. |
| [`CommandHandler`](command.go) | Tu implementación: recibe `InvokeCommand` y devuelve `InvokeResponse`. |
| [`CommandInvoker`](command.go) | Lee JSON desde un `io.Reader` (p. ej. `stdin`), valida y delega en el handler. |
| [`NewOAuth2Command`](oauth2.go) | Crea el comando raíz `oauth2` con subcomandos `get-token` y `get-url`. Úsalo directamente como root del binario. |
| [`ValidateOAuth2GetURL`](oauth2.go) | Valida argumentos requeridos para generar URL de autorización. |
| [`ValidateOAuth2GetToken`](oauth2.go) | Valida argumentos requeridos para obtener token de acceso. |
| [`GetClientCredentials`](oauth2.go) | Retorna `ClientCredentialsData` con `client_id` y `client_secret` obtenidos del endpoint de exchange usando el OTT. |
| [`ExchangeService`](exchange.go) | Realiza el HTTP POST al endpoint de exchange para intercambiar un OTT por datos. |
| [`ExchangeRequest`](exchange.go) | Cuerpo del request al exchange: `operation` y `ott`. |
| [`ExchangeReponse`](exchange.go) | Respuesta del exchange: `data` (`any`) y `message`. |
| [`ClientCredentialsData`](oauth2.go) | Estructura con `client_id` y `client_secret` retornados directamente por `GetClientCredentials`. |

## Uso rápido: invocador por stdin

1. Implementa [`CommandHandler`](command.go): método `Invoke(InvokeCommand) (InvokeResponse, error)`.
2. Crea un [`CommandInvoker`](command.go) con [`NewCommandInvoker`](command.go).
3. Llama a [`Run`](command.go) pasando el lector (habitualmente `os.Stdin`).
4. Escribe la respuesta con [`ToJSON`](json.go) (p. ej. hacia `os.Stdout`).

Ejemplo mínimo de handler:

```go
type MiHandler struct{}

func (MiHandler) Invoke(cmd grantprovider.InvokeCommand) (grantprovider.InvokeResponse, error) {
    return grantprovider.InvokeResponse{
        Result: grantprovider.Result{Success: true, Message: "ok"},
    }, nil
}
```

Entrada JSON esperada por `Run` (campos alineados con etiquetas `json` de [`InvokeCommand`](invoke.go)):

```json
{
  "command": "nombre-del-comando",
  "provider": "proveedor",
  "session_id": "id-de-sesion",
  "ott": "one-time-token",
  "exchange_endpoint": "https://exchange.example.com/api/exchange",
  "arguments": [{ "name": "clave", "value": "valor" }]
}
```

Campos requeridos: `command`, `provider`, `session_id`, `ott`, `exchange_endpoint`. El campo `arguments` es opcional. El decodificador usa [`DisallowUnknownFields`](https://pkg.go.dev/encoding/json#Decoder.DisallowUnknownFields): campos JSON desconocidos provocan error.

## Validación

- [`Validate`](validation.go) ejecuta `validate.Struct` sobre cualquier valor (típicamente un `InvokeCommand` o tus structs con etiquetas `validate`).
- Contrato de retorno: `(ValidationError, error)`:
  - Si `error != nil`: fallo distinto a errores de campo (p. ej. tipo no válido para el validador).
  - Si `error == nil` y `len(validationError.Violations) > 0`: reglas de validación incumplidas; cada ítem es un [`FieldViolation`](validation.go) (`Field`, `Namespace`, `Rule`, `Param`).
  - Si no hay violaciones: `ValidationError` vacío.

[`CommandInvoker.Run`](command.go), ante violaciones, devuelve una `InvokeResponse` con `success: false`, lista `errors` con las **reglas** fallidas (p. ej. `"required"`) y como **segundo valor de retorno** un `*ValidationError` embebido en `error` para inspección con [`errors.As`](https://pkg.go.dev/errors#As).

Para extraer detalle desde cualquier `error` de la cadena:

```go
if list, ok := grantprovider.FieldViolations(err); ok {
    for _, v := range list {
        _ = v.Namespace
        _ = v.Rule
    }
}
```

## JSON

- [`FromJSON[T]`](json.go): decodifica desde `io.Reader` con campos desconocidos rechazados.
- [`ToJSON`](json.go): codifica a `io.Writer` sin escapar HTML en strings.

## Configuración en disco

- [`GetConfigDir`](config.go): devuelve `~/.grant` (creando el directorio si no existe) o un error si falla.
- [`GetConfig`](config.go): lee `~/.grant/<fileName>` en `dest`; si el archivo no existe, escribe la configuración por defecto y rellena `dest`.

## Registro de hooks

[`Registry`](registry.go) y [`Hook`](registry.go) permiten asociar identificadores a funciones con firma:

`func(context.Context, *InvokeCommand) (*InvokeResponse, error)`

Es un contenedor; la lógica de encaminamiento por `ID` o por comando queda en tu aplicación.

## Comandos OAuth2

La librería proporciona utilidades para construir comandos OAuth2 mediante [Cobra](https://github.com/spf13/cobra).

- [`NewOAuth2Command`](oauth2.go): crea el comando raíz `oauth2` para un proveedor, agrupando `get-token` y `get-url`. Se usa directamente como root del binario (`Execute()`). Requiere que se proporcionen los comandos obligatorios `get-token` y `get-url`.

Ejemplo de implementación completa para un provider:

```go
package main

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    grantprovider "github.com/KaribuLab/grant-provider"
)

// GitHubHandler implementa CommandHandler para procesar comandos OAuth2.
type GitHubHandler struct{}

// Invoke recibe el InvokeCommand ya decodificado desde stdin.
func (h *GitHubHandler) Invoke(input grantprovider.InvokeCommand) (grantprovider.InvokeResponse, error) {
    // Obtener credenciales usando el OTT y el endpoint de exchange recibidos en el comando
    creds, err := grantprovider.GetClientCredentials(
        input.Provider,
        input.SessionID,
        input.ExchangeEndpoint,
        grantprovider.ExchangeRequest{
            Operation: grantprovider.OperationGetClientCredentials,
            OTT:       input.OTT,
        },
    )
    if err != nil {
        return grantprovider.InvokeResponse{}, fmt.Errorf("error obteniendo credenciales: %w", err)
    }
    // creds.ClientID y creds.ClientSecret ya disponibles como strings

    // Extraer arguments del input (puede ser nil)
    var arguments []grantprovider.CommandArgument
    if input.Arguments != nil {
        arguments = *input.Arguments
    }

    // Enrutar según el comando recibido
    switch input.Command {
    case "get-token":
        return h.handleGetToken(arguments)
    case "get-url":
        return h.handleGetURL(arguments)
    default:
        return grantprovider.InvokeResponse{
            Result: grantprovider.Result{
                Success: false,
                Errors:  []string{"comando desconocido"},
            },
        }, nil
    }
}

func (h *GitHubHandler) handleGetToken(arguments []grantprovider.CommandArgument) (grantprovider.InvokeResponse, error) {
    // Validar argumentos requeridos
    validationErr, err := grantprovider.ValidateOAuth2GetToken(arguments)
    if err != nil {
        return grantprovider.InvokeResponse{}, err
    }
    if len(validationErr.Violations) > 0 {
        return grantprovider.InvokeResponse{
            Result: grantprovider.Result{
                Success: false,
                Errors:  grantprovider.ListMap(validationErr.Violations, func(v grantprovider.FieldViolation) string { return v.Rule }),
            },
        }, nil
    }

    // Lógica específica del provider: intercambiar código por token con las credenciales obtenidas
    return grantprovider.InvokeResponse{
        Result: grantprovider.Result{Success: true, Message: "token obtenido"},
        Data:   map[string]any{"access_token": "gho_xxxxxxxxxxxx", "expires_in": 3600},
    }, nil
}

func (h *GitHubHandler) handleGetURL(arguments []grantprovider.CommandArgument) (grantprovider.InvokeResponse, error) {
    // Validar argumentos requeridos
    validationErr, err := grantprovider.ValidateOAuth2GetURL(arguments)
    if err != nil {
        return grantprovider.InvokeResponse{}, err
    }
    if len(validationErr.Violations) > 0 {
        return grantprovider.InvokeResponse{
            Result: grantprovider.Result{
                Success: false,
                Errors:  grantprovider.ListMap(validationErr.Violations, func(v grantprovider.FieldViolation) string { return v.Rule }),
            },
        }, nil
    }

    // Extraer valores de arguments
    argsMap := make(map[string]string)
    for _, arg := range arguments {
        argsMap[arg.Name] = arg.Value
    }

    // Construir URL de autorización
    authURL := fmt.Sprintf(
        "https://github.com/login/oauth/authorize?response_type=%s&client_id=%s&redirect_uri=%s&scope=%s&state=%s",
        argsMap["response_type"],
        argsMap["client_id"],
        argsMap["redirect_uri"],
        argsMap["scope"],
        argsMap["state"],
    )

    return grantprovider.InvokeResponse{
        Result: grantprovider.Result{Success: true, Message: "URL generada"},
        Data:   map[string]string{"auth_url": authURL},
    }, nil
}

func main() {
    // Crear comando get-token: lee JSON desde stdin, delega al handler
    tokenCmd := &cobra.Command{
        Use:   "get-token",
        Short: "Obtiene un token de acceso",
        RunE: func(cmd *cobra.Command, args []string) error {
            handler := &GitHubHandler{}
            invoker := grantprovider.NewCommandInvoker(handler)
            response, err := invoker.Run(os.Stdin)
            if err != nil {
                // Si hay error de validación, response ya contiene los detalles
                _ = grantprovider.ToJSON(response, os.Stdout)
                return err
            }
            return grantprovider.ToJSON(response, os.Stdout)
        },
    }

    // Crear comando get-url: lee JSON desde stdin, delega al handler
    urlCmd := &cobra.Command{
        Use:   "get-url",
        Short: "Genera URL de autorización",
        RunE: func(cmd *cobra.Command, args []string) error {
            handler := &GitHubHandler{}
            invoker := grantprovider.NewCommandInvoker(handler)
            response, err := invoker.Run(os.Stdin)
            if err != nil {
                _ = grantprovider.ToJSON(response, os.Stdout)
                return err
            }
            return grantprovider.ToJSON(response, os.Stdout)
        },
    }

    // NewOAuth2Command retorna el root command del binario
    // Invocación: ./grant-github get-token  o  ./grant-github get-url
    rootCmd, err := grantprovider.NewOAuth2Command("github", grantprovider.OAuth2Commands{
        "get-token": tokenCmd,
        "get-url":   urlCmd,
    })
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

**Flujo de datos:**

```
stdin (JSON) → CommandInvoker.Run → FromJSON → CommandHandler.Invoke → ToJSON → stdout
```

Ejemplo de entrada JSON esperada:

```json
{
  "command": "get-token",
  "provider": "github",
  "session_id": "sess-123",
  "ott": "one-time-token-abc123",
  "exchange_endpoint": "https://exchange.example.com/api/exchange",
  "arguments": [
    {"name": "code", "value": "abc123def456"}
  ]
}
```

**Patrón recomendado para múltiples providers:**

Separar el handler del comando para facilitar testing y reuso:

```go
// github/handler.go
package github

import (
    "fmt"
    "net/http"
    "net/url"

    grantprovider "github.com/KaribuLab/grant-provider"
)

type Handler struct{}

func (h *Handler) Invoke(input grantprovider.InvokeCommand) (grantprovider.InvokeResponse, error) {
    // Obtener credenciales del cliente vía exchange
    creds, err := grantprovider.GetClientCredentials(
        input.Provider,
        input.SessionID,
        input.ExchangeEndpoint,
        grantprovider.ExchangeRequest{
            Operation: grantprovider.OperationGetClientCredentials,
            OTT:       input.OTT,
        },
    )
    if err != nil {
        return grantprovider.InvokeResponse{}, fmt.Errorf("error obteniendo credenciales: %w", err)
    }
    // creds.ClientID y creds.ClientSecret ya disponibles como strings

    var arguments []grantprovider.CommandArgument
    if input.Arguments != nil {
        arguments = *input.Arguments
    }

    switch input.Command {
    case "get-token":
        return h.getToken(arguments)
    case "get-url":
        return h.getURL(arguments)
    default:
        return grantprovider.InvokeResponse{
            Result: grantprovider.Result{
                Success: false,
                Errors:  []string{"comando no soportado: " + input.Command},
            },
        }, nil
    }
}

func (h *Handler) getToken(arguments []grantprovider.CommandArgument) (grantprovider.InvokeResponse, error) {
    validationErr, err := grantprovider.ValidateOAuth2GetToken(arguments)
    if err != nil || len(validationErr.Violations) > 0 {
        return grantprovider.InvokeResponse{
            Result: grantprovider.Result{
                Success: false,
                Errors:  []string{"argumentos inválidos"},
            },
        }, nil
    }

    // Lógica específica del provider usando las credenciales obtenidas vía exchange...
    return grantprovider.InvokeResponse{
        Result: grantprovider.Result{Success: true},
        Data:   map[string]any{"access_token": "token"},
    }, nil
}

func (h *Handler) getURL(arguments []grantprovider.CommandArgument) (grantprovider.InvokeResponse, error) {
    validationErr, err := grantprovider.ValidateOAuth2GetURL(arguments)
    if err != nil || len(validationErr.Violations) > 0 {
        return grantprovider.InvokeResponse{
            Result: grantprovider.Result{
                Success: false,
                Errors:  []string{"argumentos inválidos"},
            },
        }, nil
    }

    // Construir URL...
    return grantprovider.InvokeResponse{
        Result: grantprovider.Result{Success: true},
        Data:   map[string]string{"auth_url": "https://..."},
    }, nil
}
```

Y en el punto de entrada del provider:

```go
// github/main.go
package main

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    grantprovider "github.com/KaribuLab/grant-provider"
    "github.com/KaribuLab/provider-github/github"
)

func main() {
    handler := &github.Handler{
        ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
        ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
    }

    commands := grantprovider.OAuth2Commands{
        "get-token": buildCmd(handler, "get-token"),
        "get-url":   buildCmd(handler, "get-url"),
    }

    // NewOAuth2Command retorna el root command del binario
    // Invocación: ./grant-github get-token  o  ./grant-github get-url
    rootCmd, err := grantprovider.NewOAuth2Command("github", commands)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

func buildCmd(handler *github.Handler, commandName string) *cobra.Command {
    return &cobra.Command{
        Use:   commandName,
        Short: fmt.Sprintf("OAuth2 %s para GitHub", commandName),
        RunE: func(cmd *cobra.Command, args []string) error {
            invoker := grantprovider.NewCommandInvoker(handler)
            response, err := invoker.Run(os.Stdin)
            _ = grantprovider.ToJSON(response, os.Stdout)
            return err
        },
    }
}
```

Si falta algún comando requerido, `NewOAuth2Command` retorna error indicando cuáles faltan.

> **Cómo Cobra resuelve los comandos:** En Cobra, el campo `Use` del root command es solo texto para el `--help`; la ruta de invocación real siempre parte del **nombre del binario** (`os.Args[0]`). Por eso, aunque el root tenga `Use: "oauth2"`, los subcomandos se invocan directamente después del binario, **sin** repetir `oauth2`:
>
> ```bash
> # Correcto: ./binario subcomando
> echo '{...}' | ./grant-github get-url
> echo '{...}' | ./grant-github get-token
>
> # Incorrecto: ./binario root subcomando — produce "unknown command 'oauth2' for 'oauth2'"
> echo '{...}' | ./grant-github oauth2 get-url
> ```
>
> Si necesitas el prefijo `oauth2` en la invocación (`./grant-github oauth2 get-url`), crea un root command propio y agrega el resultado de `NewOAuth2Command` con `rootCmd.AddCommand(oauth2Cmd)`.

### Validación de argumentos OAuth2

La librería proporciona validadores reutilizables para verificar que los argumentos de los comandos OAuth2 cumplan con los requisitos mínimos:

- [`ValidateOAuth2GetURL`](oauth2.go): valida que estén presentes los parámetros obligatorios para generar una URL de autorización.
  - Requiere: `response_type`, `client_id`, `redirect_uri`, `scope`, `state`.
- [`ValidateOAuth2GetToken`](oauth2.go): valida que esté presente el parámetro obligatorio para intercambiar el código por un token.
  - Requiere: `code`.
  - Nota: `grant_type` no se valida porque se asume fijo como `code` en la lógica del proveedor.

Ambas funciones retornan [`ValidationError`](validation.go) con la lista de violaciones para campos faltantes:

```go
arguments := []grantprovider.CommandArgument{
    {Name: "response_type", Value: "code"},
    {Name: "client_id", Value: "my-client-id"},
    {Name: "redirect_uri", Value: "https://example.com/callback"},
    {Name: "scope", Value: "openid email"},
    {Name: "state", Value: "abc123"},
}

validationErr, err := grantprovider.ValidateOAuth2GetURL(arguments)
if err != nil {
    // Error interno del validador
    return err
}
if len(validationErr.Violations) > 0 {
    // Procesar violaciones: cada una tiene Field, Namespace, Rule
    for _, v := range validationErr.Violations {
        fmt.Printf("Campo %s: regla %s incumplida\n", v.Field, v.Rule)
    }
}
```

## Obtención de credenciales del cliente

El proveedor recibe en cada `InvokeCommand` un **OTT** (`ott`) y un **endpoint de exchange** (`exchange_endpoint`). Estos permiten recuperar de forma segura las credenciales del cliente (`client_id` y `client_secret`) llamando a [`GetClientCredentials`](oauth2.go).

```go
func (h *MyHandler) Invoke(input grantprovider.InvokeCommand) (grantprovider.InvokeResponse, error) {
    // Obtener credenciales del cliente usando el OTT recibido en el comando
    creds, err := grantprovider.GetClientCredentials(
        input.Provider,
        input.SessionID,
        input.ExchangeEndpoint,
        grantprovider.ExchangeRequest{
            Operation: grantprovider.OperationGetClientCredentials,
            OTT:       input.OTT,
        },
    )
    if err != nil {
        return grantprovider.InvokeResponse{}, fmt.Errorf("error obteniendo credenciales: %w", err)
    }

    // creds.ClientID y creds.ClientSecret disponibles como strings para usar en las llamadas OAuth2
    _ = creds.ClientID
    _ = creds.ClientSecret
    return grantprovider.InvokeResponse{
        Result: grantprovider.Result{Success: true, Message: "operación completada"},
    }, nil
}
```

### Tipos del exchange

| Tipo | Descripción |
|------|-------------|
| [`ExchangeRequest`](exchange.go) | Body del request: `operation` (usar `OperationGetClientCredentials`) y `ott`. |
| [`ExchangeReponse`](exchange.go) | Respuesta: `data` (`any`) con las credenciales y `message`. |
| [`ClientCredentialsData`](oauth2.go) | Estructura con `client_id` y `client_secret` retornados directamente por `GetClientCredentials`. |

El endpoint construido internamente sigue el patrón:
```
{exchange_endpoint}/{provider}/{session_id}
```

## Dependencias directas relevantes

- [`github.com/go-playground/validator/v10`](https://github.com/go-playground/validator) — validación por etiquetas.
- [`github.com/spf13/cobra`](https://github.com/spf13/cobra) — framework para comandos CLI (usado en [`NewOAuth2Command`](oauth2.go)).

## Documentación en código

```bash
go doc -all github.com/KaribuLab/grant-provider
```
