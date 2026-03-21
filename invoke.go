package grantprovider

// Result resume el éxito o fallo lógico de una invocación y mensajes o errores asociados.
type Result struct {
	Success bool     `json:"success" validate:"required"`
	Message string   `json:"message,omitempty" validate:"omitempty"`
	Errors  []string `json:"errors,omitempty" validate:"omitempty"`
}

// CommandArgument representa un par nombre/valor opcional para enriquecer el comando.
type CommandArgument struct {
	Name  string `json:"name" validate:"required"`
	Value string `json:"value" validate:"required"`
}

// InvokeCommand es el cuerpo de entrada decodificado desde JSON para un comando.
type InvokeCommand struct {
	Arguments *[]CommandArgument `json:"arguments,omitempty" validate:"omitempty"`
	Command   string             `json:"command" validate:"required"`
	Provider  string             `json:"provider" validate:"required"`
	SessionID string             `json:"session_id" validate:"required"`
}

// InvokeResponse representa la salida de un comando. Data admite cualquier forma
// (p. ej. *GetAccessTokenData u otro struct concreto según el comando).
type InvokeResponse struct {
	Result         `json:"result"`
	Data           any             `json:"data,omitempty" validate:"omitempty"`
	AdditionalData *map[string]any `json:"additional_data,omitempty" validate:"omitempty"`
}
