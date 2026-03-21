package grantprovider

import (
	"strings"
	"testing"
)

func TestNewOAuth2Command_Success(t *testing.T) {
	provider := "github"
	commands := OAuth2Commands{
		"get-token": {Use: "get-token"},
		"get-url":   {Use: "get-url"},
	}

	root, err := NewOAuth2Command(provider, commands)
	if err != nil {
		t.Fatalf("se esperaba nil error, se obtuvo: %v", err)
	}
	if root == nil {
		t.Fatal("se esperaba comando raiz no nil")
	}
	if root.Use != "oauth2" {
		t.Fatalf("Use inesperado, esperado oauth2, obtenido %s", root.Use)
	}
	expectedShort := "Procesa operaciones de oauth2 para el proveedor github"
	if root.Short != expectedShort {
		t.Fatalf("Short inesperado, esperado %q, obtenido %q", expectedShort, root.Short)
	}

	children := root.Commands()
	if len(children) != 2 {
		t.Fatalf("se esperaban 2 subcomandos, se obtuvieron %d", len(children))
	}

	got := map[string]bool{}
	for _, cmd := range children {
		if cmd == nil {
			t.Fatal("se encontro subcomando nil")
		}
		got[cmd.Use] = true
	}
	if !got["get-token"] || !got["get-url"] {
		t.Fatalf("subcomandos inesperados: %+v", got)
	}
}

func TestNewOAuth2Command_MissingRequiredCommand(t *testing.T) {
	commands := OAuth2Commands{
		"get-token": {Use: "get-token"},
	}

	root, err := NewOAuth2Command("github", commands)
	if err == nil {
		t.Fatal("se esperaba error por comando requerido faltante")
	}
	if root != nil {
		t.Fatal("se esperaba root nil cuando faltan comandos requeridos")
	}
	if !strings.Contains(err.Error(), "comandos requeridos no encontrados") {
		t.Fatalf("mensaje de error inesperado: %v", err)
	}
	if !strings.Contains(err.Error(), "get-url") {
		t.Fatalf("se esperaba que el error mencione get-url, se obtuvo: %v", err)
	}
}

func TestNewOAuth2Command_MissingBothRequiredCommands(t *testing.T) {
	root, err := NewOAuth2Command("github", OAuth2Commands{})
	if err == nil {
		t.Fatal("se esperaba error por comandos requeridos faltantes")
	}
	if root != nil {
		t.Fatal("se esperaba root nil cuando faltan comandos requeridos")
	}
	// El orden del map no es estable, por eso validamos presencia sin orden.
	if !strings.Contains(err.Error(), "get-token") || !strings.Contains(err.Error(), "get-url") {
		t.Fatalf("se esperaban ambos comandos en el error, se obtuvo: %v", err)
	}
}

// ========== Pruebas para ValidateOAuth2GetURL ==========

func TestValidateOAuth2GetURL_Success(t *testing.T) {
	arguments := []CommandArgument{
		{Name: "response_type", Value: "code"},
		{Name: "client_id", Value: "my-client-id"},
		{Name: "redirect_uri", Value: "https://example.com/callback"},
		{Name: "scope", Value: "openid profile email"},
		{Name: "state", Value: "random-state-123"},
	}

	validationErr, err := ValidateOAuth2GetURL(arguments)
	if err != nil {
		t.Fatalf("se esperaba error nil de validación, se obtuvo: %v", err)
	}
	if len(validationErr.Violations) > 0 {
		t.Fatalf("se esperaba sin violaciones, se obtuvieron: %+v", validationErr.Violations)
	}
}

func TestValidateOAuth2GetURL_MissingSingleParam(t *testing.T) {
	arguments := []CommandArgument{
		{Name: "response_type", Value: "code"},
		{Name: "client_id", Value: "my-client-id"},
		{Name: "redirect_uri", Value: "https://example.com/callback"},
		{Name: "scope", Value: "openid profile email"},
		// Falta "state"
	}

	validationErr, err := ValidateOAuth2GetURL(arguments)
	if err != nil {
		t.Fatalf("se esperaba error nil de validación, se obtuvo: %v", err)
	}
	if len(validationErr.Violations) != 1 {
		t.Fatalf("se esperaba 1 violación, se obtuvieron: %d", len(validationErr.Violations))
	}
	if validationErr.Violations[0].Field != "state" {
		t.Fatalf("se esperaba violación en campo 'state', se obtuvo: %s", validationErr.Violations[0].Field)
	}
	if validationErr.Violations[0].Rule != "required" {
		t.Fatalf("se esperaba regla 'required', se obtuvo: %s", validationErr.Violations[0].Rule)
	}
}

func TestValidateOAuth2GetURL_MissingMultipleParams(t *testing.T) {
	arguments := []CommandArgument{
		{Name: "response_type", Value: "code"},
		{Name: "client_id", Value: "my-client-id"},
		// Faltan "redirect_uri", "scope", "state"
	}

	validationErr, err := ValidateOAuth2GetURL(arguments)
	if err != nil {
		t.Fatalf("se esperaba error nil de validación, se obtuvo: %v", err)
	}
	if len(validationErr.Violations) != 3 {
		t.Fatalf("se esperaban 3 violaciones, se obtuvieron: %d", len(validationErr.Violations))
	}

	// Verificar que todos los campos faltantes están en las violaciones
	missingFields := make(map[string]bool)
	for _, v := range validationErr.Violations {
		missingFields[v.Field] = true
	}
	if !missingFields["redirect_uri"] || !missingFields["scope"] || !missingFields["state"] {
		t.Fatalf("violaciones inesperadas: %+v", validationErr.Violations)
	}
}

func TestValidateOAuth2GetURL_EmptyArguments(t *testing.T) {
	arguments := []CommandArgument{}

	validationErr, err := ValidateOAuth2GetURL(arguments)
	if err != nil {
		t.Fatalf("se esperaba error nil de validación, se obtuvo: %v", err)
	}
	if len(validationErr.Violations) != 5 {
		t.Fatalf("se esperaban 5 violaciones (todos los campos faltantes), se obtuvieron: %d", len(validationErr.Violations))
	}
}

// ========== Pruebas para ValidateOAuth2GetToken ==========

func TestValidateOAuth2GetToken_Success(t *testing.T) {
	arguments := []CommandArgument{
		{Name: "code", Value: "auth-code-abc123"},
	}

	validationErr, err := ValidateOAuth2GetToken(arguments)
	if err != nil {
		t.Fatalf("se esperaba error nil de validación, se obtuvo: %v", err)
	}
	if len(validationErr.Violations) > 0 {
		t.Fatalf("se esperaba sin violaciones, se obtuvieron: %+v", validationErr.Violations)
	}
}

func TestValidateOAuth2GetToken_WithExtraParams(t *testing.T) {
	// El validador no debe rechazar parámetros adicionales como grant_type
	arguments := []CommandArgument{
		{Name: "code", Value: "auth-code-abc123"},
		{Name: "grant_type", Value: "code"},
		{Name: "redirect_uri", Value: "https://example.com/callback"},
	}

	validationErr, err := ValidateOAuth2GetToken(arguments)
	if err != nil {
		t.Fatalf("se esperaba error nil de validación, se obtuvo: %v", err)
	}
	if len(validationErr.Violations) > 0 {
		t.Fatalf("se esperaba sin violaciones (grant_type es opcional), se obtuvieron: %+v", validationErr.Violations)
	}
}

func TestValidateOAuth2GetToken_MissingCode(t *testing.T) {
	arguments := []CommandArgument{
		{Name: "redirect_uri", Value: "https://example.com/callback"},
	}

	validationErr, err := ValidateOAuth2GetToken(arguments)
	if err != nil {
		t.Fatalf("se esperaba error nil de validación, se obtuvo: %v", err)
	}
	if len(validationErr.Violations) != 1 {
		t.Fatalf("se esperaba 1 violación, se obtuvieron: %d", len(validationErr.Violations))
	}
	if validationErr.Violations[0].Field != "code" {
		t.Fatalf("se esperaba violación en campo 'code', se obtuvo: %s", validationErr.Violations[0].Field)
	}
	if validationErr.Violations[0].Rule != "required" {
		t.Fatalf("se esperaba regla 'required', se obtuvo: %s", validationErr.Violations[0].Rule)
	}
}

func TestValidateOAuth2GetToken_EmptyArguments(t *testing.T) {
	arguments := []CommandArgument{}

	validationErr, err := ValidateOAuth2GetToken(arguments)
	if err != nil {
		t.Fatalf("se esperaba error nil de validación, se obtuvo: %v", err)
	}
	if len(validationErr.Violations) != 1 {
		t.Fatalf("se esperaba 1 violación (code faltante), se obtuvieron: %d", len(validationErr.Violations))
	}
	if validationErr.Violations[0].Field != "code" {
		t.Fatalf("se esperaba violación en campo 'code', se obtuvo: %s", validationErr.Violations[0].Field)
	}
}


