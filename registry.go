package grantprovider

import "context"

// Hook asocia un identificador con un manejador que puede invocar comandos con contexto.
type Hook struct {
	// ID identifica el hook (uso definido por la aplicación).
	ID string
	// Handler ejecuta la lógica del hook con contexto y comando.
	Handler func(ctx context.Context, req *InvokeCommand) (*InvokeResponse, error)
}

// Registry agrupa hooks; el enrutamiento por ID o comando queda fuera de este paquete.
type Registry struct {
	// Hooks lista registrada; no está ordenada ni deduplicada por este tipo.
	Hooks []Hook
}
