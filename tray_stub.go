//go:build !windows

package main

type trayController interface {
	Init(app *App)
	Dispose()
}

func newTrayController() trayController {
	return nil
}
