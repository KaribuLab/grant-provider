package grantprovider

import (
	"bytes"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

const (
	ContentTypeJSON = "application/json"
)

const (
	OperationGetClientCredentials = "client_credentials"
)

// ExchangeRequest es el cuerpo enviado al endpoint de exchange para intercambiar un OTT.
type ExchangeRequest struct {
	// Operation identifica la operación solicitada. Usar OperationGetClientCredentials para credenciales.
	Operation string `json:"operation" validate:"required"`
	// OTT es el one-time token recibido en InvokeCommand.OTT.
	OTT string `json:"ott" validate:"required"`
}

// ExchangeReponse es la respuesta genérica del endpoint de exchange.
// El campo Data contiene el payload específico de cada operación.
type ExchangeReponse struct {
	Data    any    `json:"data" validate:"required"`
	Message string `json:"message" validate:"required"`
}

// ExchangeFetcher define el contrato para ejecutar un intercambio con el endpoint de exchange.
// Implementar esta interfaz permite sustituir ExchangeFetcherService por un mock en tests.
type ExchangeFetcher interface {
	Execute(ExchangeRequest) (ExchangeReponse, error)
}

// ExchangeFetcherService implementa ExchangeFetcher realizando un HTTP POST al endpoint de exchange.
// La URL final sigue el patrón: {ExchangeEndpoint}/{Provider}/{SessionID}.
type ExchangeFetcherService struct {
	Provider         string
	SessionID        string
	ExchangeEndpoint string
}

// Execute realiza el HTTP POST al endpoint de exchange y retorna la respuesta deserializada.
func (e *ExchangeFetcherService) Execute(exchangeRequest ExchangeRequest) (ExchangeReponse, error) {
	endpoint := strings.Join([]string{
		e.ExchangeEndpoint,
		url.QueryEscape(e.Provider),
		url.QueryEscape(e.SessionID),
	}, "/")
	body := new(bytes.Buffer)
	err := ToJSON(body, exchangeRequest)
	if err != nil {
		return ExchangeReponse{}, errors.New("error intentando transormar a JSON request")
	}
	response, err := http.Post(
		endpoint,
		ContentTypeJSON,
		body,
	)
	if err != nil {
		return ExchangeReponse{}, err
	}
	exchangeResponse, err := FromJSON[ExchangeReponse](response.Body)
	return exchangeResponse, nil
}
