package grantprovider

import (
	"errors"
	"strings"
	"testing"
)

// testData es un struct local para probar el campo Data de InvokeResponse.
type testData struct {
	Token string `json:"token"`
}

// mockHandler implementa CommandHandler para tests.
type mockHandler struct {
	resp InvokeResponse
	err  error
}

func (m *mockHandler) Invoke(input InvokeCommand) (InvokeResponse, error) {
	return m.resp, m.err
}

func TestCommandInvokerRun(t *testing.T) {
	t.Run("flujo exitoso", func(t *testing.T) {
		jsonInput := `{"command":"test","provider":"mock","session_id":"123"}`
		handler := &mockHandler{
			resp: InvokeResponse{
				Result: Result{Success: true, Message: "ok"},
			},
		}
		invoker := NewCommandInvoker(handler)

		resp, err := invoker.Run(strings.NewReader(jsonInput))

		if err != nil {
			t.Fatalf("esperaba nil error, obtuve: %v", err)
		}
		if !resp.Success {
			t.Errorf("esperaba Success true, obtuve false")
		}
		if resp.Message != "ok" {
			t.Errorf("esperaba Message 'ok', obtuve: %s", resp.Message)
		}
	})

	t.Run("JSON invalido", func(t *testing.T) {
		jsonInput := `{"command":"test", invalido}`
		handler := &mockHandler{}
		invoker := NewCommandInvoker(handler)

		resp, err := invoker.Run(strings.NewReader(jsonInput))

		if err == nil {
			t.Fatal("esperaba error por JSON invalido, obtuve nil")
		}
		if resp.Success {
			t.Error("esperaba respuesta vacia (Success false)")
		}
	})

	t.Run("campo requerido faltante", func(t *testing.T) {
		// Falta el campo "command" que es required
		jsonInput := `{"provider":"mock","session_id":"123"}`
		handler := &mockHandler{}
		invoker := NewCommandInvoker(handler)

		resp, err := invoker.Run(strings.NewReader(jsonInput))

		if err == nil {
			t.Fatal("esperaba error de validacion, obtuve nil")
		}
		if resp.Success {
			t.Error("esperaba Success false por validacion fallida")
		}

		// Verificar que es ValidationError
		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Errorf("esperaba *ValidationError, obtuve: %T", err)
		} else {
			if len(valErr.Violations) == 0 {
				t.Error("esperaba al menos una violacion")
			}
			foundRequired := false
			for _, v := range valErr.Violations {
				if v.Rule == "required" {
					foundRequired = true
					break
				}
			}
			if !foundRequired {
				t.Error("esperaba violacion 'required' en las violaciones")
			}
		}

		// Verificar que result.Errors contiene "required"
		if len(resp.Errors) == 0 {
			t.Error("esperaba resp.Errors con al menos un error")
		}
		foundInErrors := false
		for _, e := range resp.Errors {
			if e == "required" {
				foundInErrors = true
				break
			}
		}
		if !foundInErrors {
			t.Errorf("esperaba 'required' en resp.Errors, obtuve: %v", resp.Errors)
		}
	})

	t.Run("error del handler", func(t *testing.T) {
		jsonInput := `{"command":"test","provider":"mock","session_id":"123"}`
		handler := &mockHandler{
			err: errors.New("error simulado del handler"),
		}
		invoker := NewCommandInvoker(handler)

		resp, err := invoker.Run(strings.NewReader(jsonInput))

		if err == nil {
			t.Fatal("esperaba error del handler, obtuve nil")
		}
		if err.Error() != "error simulado del handler" {
			t.Errorf("mensaje de error inesperado: %v", err)
		}
		if resp.Success {
			t.Error("esperaba respuesta vacia ante error del handler")
		}
	})

	t.Run("respuesta con Data", func(t *testing.T) {
		jsonInput := `{"command":"test","provider":"mock","session_id":"123"}`
		data := &testData{Token: "abc123"}
		handler := &mockHandler{
			resp: InvokeResponse{
				Result: Result{Success: true},
				Data:   data,
			},
		}
		invoker := NewCommandInvoker(handler)

		resp, err := invoker.Run(strings.NewReader(jsonInput))

		if err != nil {
			t.Fatalf("esperaba nil error, obtuve: %v", err)
		}
		if !resp.Success {
			t.Error("esperaba Success true")
		}

		// Verificar que Data se copió correctamente
		respData, ok := resp.Data.(*testData)
		if !ok {
			t.Fatalf("esperaba Data tipo *testData, obtuve: %T", resp.Data)
		}
		if respData.Token != "abc123" {
			t.Errorf("esperaba Token 'abc123', obtuve: %s", respData.Token)
		}
	})
}
