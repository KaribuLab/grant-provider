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
| [`InvokeCommand`](invoke.go) | Entrada: comando, proveedor, sesión y argumentos opcionales. |
| [`InvokeResponse`](invoke.go) | Salida: `result` embebido, `data` opcional (`any`) y `additional_data` opcional. |
| [`CommandHandler`](command.go) | Tu implementación: recibe `InvokeCommand` y devuelve `InvokeResponse`. |
| [`CommandInvoker`](command.go) | Lee JSON desde un `io.Reader` (p. ej. `stdin`), valida y delega en el handler. |

## Uso rápido: invocador por stdin

1. Implementa [`CommandHandler`](command.go): método `Invoke(InvokeCommand) (InvokeResponse, error)`.
2. Crea un [`CommandInvoker`](command.go) con [`NewCommandInvoker`](command.go).
3. Llama a [`Run`](command.go) pasando el lector (habitualmente `os.Stdin`).
4. Escribe la respuesta con [`ToJSON`](json.go) (p. ej. hacia `os.Stdout`).

Ejemplo mínimo de handler:

```go
type MiHandler struct{}

func (MiHandler) Invoke(cmd grantprovider.InvokeCommand) (grantprovider.InvokeResponse, error) {
    // Ejemplo: devolver datos concretos en Data (cualquier tipo vía any)
    return grantprovider.InvokeResponse{
        Result: grantprovider.Result{Success: true, Message: "ok"},
        Data: &grantprovider.GetAccessTokenData{
            AccessToken: "…", RefreshToken: "…", ExpiresIn: 3600,
        },
    }, nil
}
```

Entrada JSON esperada por `Run` (campos alineados con etiquetas `json` de [`InvokeCommand`](invoke.go)):

```json
{
  "command": "nombre-del-comando",
  "provider": "proveedor",
  "session_id": "id-de-sesion",
  "arguments": [{ "name": "clave", "value": "valor" }]
}
```

`arguments` es opcional. El decodificador usa [`DisallowUnknownFields`](https://pkg.go.dev/encoding/json#Decoder.DisallowUnknownFields): campos JSON desconocidos provocan error.

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

- [`GetConfigDir`](config.go): devuelve `~/.grant`, creando el directorio si no existe.
- [`GetConfig`](config.go): lee `~/.grant/<fileName>` en `dest`; si el archivo no existe, escribe la configuración por defecto y rellena `dest`.

**Nota:** [`GetConfigDir`](config.go) puede llamar a `log.Fatal` si no puede obtener el directorio home o crear `~/.grant`. [`GetConfig`](config.go) usa `log` para avisos/trazas según el nivel configurado en logrus.

## Registro de hooks

[`Registry`](registry.go) y [`Hook`](registry.go) permiten asociar identificadores a funciones con firma:

`func(context.Context, *InvokeCommand) (*InvokeResponse, error)`

Es un contenedor; la lógica de encaminamiento por `ID` o por comando queda en tu aplicación.

## Tipos de datos de ejemplo

- [`GetAccessTokenData`](oauth2.go): ejemplo de payload para tokens OAuth2 en `InvokeResponse.Data` (puedes definir otros structs y asignarlos a `Data` como `any`).

## Dependencias directas relevantes

- [`github.com/go-playground/validator/v10`](https://github.com/go-playground/validator) — validación por etiquetas.
- [`github.com/sirupsen/logrus`](https://github.com/sirupsen/logrus) — logging en configuración.

## Documentación en código

```bash
go doc -all github.com/KaribuLab/grant-provider
```
