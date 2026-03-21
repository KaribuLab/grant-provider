package grantprovider

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// FieldViolation describe un fallo de validación en un campo concreto.
type FieldViolation struct {
	Field     string // nombre del campo (p. ej. "Command")
	Namespace string // ruta con structs anidados (p. ej. "InvokeCommand.Command")
	Rule      string // regla que falló (p. ej. "required")
	Param     string // parámetro de la regla (p. ej. "10" para max=10)
}

// ValidationError agrupa los fallos de validate.Struct para inspección programática.
type ValidationError struct {
	Violations []FieldViolation
	cause      error // error original del validador (cadena con errors.Unwrap)
}

// Error implementa error y resume las violaciones en un solo mensaje.
func (e *ValidationError) Error() string {
	if len(e.Violations) == 0 {
		return "validation error"
	}
	var b strings.Builder
	b.WriteString("validation error:")
	for _, v := range e.Violations {
		fmt.Fprintf(&b, " %s: %s", v.Namespace, v.Rule)
		if v.Param != "" {
			fmt.Fprintf(&b, "=%s", v.Param)
		}
		b.WriteString(";")
	}
	return strings.TrimSuffix(b.String(), ";")
}

// Unwrap devuelve el error original del validador para errors.Is / errors.As.
func (e *ValidationError) Unwrap() error { return e.cause }

// FieldViolations devuelve el detalle de fallos si err (o su cadena Unwrap) es *ValidationError
// o validator.ValidationErrors. Si no hay detalle, ok es false.
func FieldViolations(err error) ([]FieldViolation, bool) {
	var ve *ValidationError
	if errors.As(err, &ve) && len(ve.Violations) > 0 {
		return ve.Violations, true
	}
	var raw validator.ValidationErrors
	if errors.As(err, &raw) {
		return violationsFrom(raw), true
	}
	return nil, false
}

func violationsFrom(err validator.ValidationErrors) []FieldViolation {
	out := make([]FieldViolation, 0, len(err))
	for _, fe := range err {
		out = append(out, FieldViolation{
			Field:     fe.Field(),
			Namespace: fe.Namespace(),
			Rule:      fe.Tag(),
			Param:     fe.Param(),
		})
	}
	return out
}

// Validate aplica validate.Struct a data según etiquetas validate. Si la
// validación falla por reglas de campo, devuelve ValidationError con Violations
// rellenado y error nil. Si err != nil, es un fallo distinto (p. ej. tipo no válido).
func Validate[T any](data T) (ValidationError, error) {
	if err := validate.Struct(data); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			return ValidationError{
				Violations: violationsFrom(validationErrors),
				cause:      err,
			}, nil
		}
		return ValidationError{}, err
	}
	return ValidationError{}, nil
}
