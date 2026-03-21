// Package grantprovider ofrece tipos y utilidades para invocar comandos con
// entrada/salida JSON, validación con etiquetas struct y configuración en
// ~/.grant.
//
// Flujo típico: decodificar un [InvokeCommand] desde JSON ([FromJSON]),
// validar con [Validate], ejecutar un [CommandHandler] vía [CommandInvoker.Run]
// y serializar la [InvokeResponse] con [ToJSON].
//
// La validación usa github.com/go-playground/validator/v10; los fallos por campo
// se exponen como [ValidationError] y [FieldViolations].
package grantprovider
