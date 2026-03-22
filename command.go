package grantprovider

import (
	"io"

	"github.com/go-playground/validator/v10"
)

// CommandHandler define la ejecución de un comando invocado vía JSON.
type CommandHandler interface {
	// Invoke procesa input y devuelve la respuesta o un error de ejecución.
	Invoke(input InvokeCommand) (InvokeResponse, error)
}

// CommandInvoker orquesta decodificación JSON, validación y delegación al handler.
type CommandInvoker struct {
	handler  CommandHandler
	validate *validator.Validate
}

// NewCommandInvoker crea un invocador que usa el handler dado.
func NewCommandInvoker(h CommandHandler) *CommandInvoker {
	return &CommandInvoker{
		handler:  h,
		validate: validator.New(),
	}
}

// Run lee un InvokeCommand desde stdin (u otro reader), valida con etiquetas
// validate y, si todo es correcto, llama a Invoke del handler. Si hay violaciones
// de validación, devuelve una respuesta con Success false y un segundo valor
// error que puede ser *ValidationError (usar errors.As).
func (ci *CommandInvoker) Run(stdin io.Reader) (InvokeResponse, error) {
	input, err := FromJSON[InvokeCommand](stdin)
	if err != nil {
		return InvokeResponse{}, err
	}

	validationError, err := Validate(input)
	if err != nil {
		return InvokeResponse{}, err
	}

	if len(validationError.Violations) > 0 {
		return InvokeResponse{
			Result: Result{
				Success: false,
				Errors: ListMap(validationError.Violations, func(v FieldViolation) string {
					return v.Rule
				}),
			},
		}, &validationError
	}

	output, err := ci.handler.Invoke(input)
	if err != nil {
		return InvokeResponse{}, err
	}

	return InvokeResponse{
		Result:         output.Result,
		Data:           output.Data,
		AdditionalData: output.AdditionalData,
	}, nil
}
