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


