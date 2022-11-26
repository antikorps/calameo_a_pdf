package main

import (
	"context"
	"errors"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) GoDescargar(url, carpetaDestino string, descargasSimultaneas, tiempoEspera int) (string, error) {
	manejador := NuevoManejador(url, carpetaDestino, descargasSimultaneas, tiempoEspera)
	manejador.Validar()
	manejador.Descargar()
	manejador.CrearPDF()
	manejador.BorrarArchivos()
	return manejador.GenerarRespuesta()
}

func (a *App) GoRutaDestino() (string, error) {
	carpetaDestino, carpetaDestinoError := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{})
	if carpetaDestinoError != nil {
		return "", carpetaDestinoError
	}
	if carpetaDestino == "" {
		return "", errors.New("no se ha seleccionado ninguna carpeta")
	}
	return carpetaDestino, nil
}
