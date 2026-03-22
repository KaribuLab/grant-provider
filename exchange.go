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

type ExchangeRequest struct {
	Operation string `json:"operation" validate:"required"`
	OTT       string `json:"ott" validate:"required"`
}

type ExchangeReponse struct {
	Data    any    `json:"data" validate:"required"`
	Message string `json:"message" validate:"required"`
}

type ExchangeService struct {
	Provider         string
	SessionID        string
	ExchangeEndpoint string
}

func (e *ExchangeService) Execute(exchangeRequest ExchangeRequest) (ExchangeReponse, error) {
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
