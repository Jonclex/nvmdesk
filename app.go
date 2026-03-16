package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct - main application structure exposed to frontend
type App struct {
	ctx        context.Context
	nvmService *NvmService
	tray       trayController
	isQuitting bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		nvmService: NewNvmService(),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if a.tray != nil {
		a.tray.Init(a)
	}
}

func (a *App) beforeClose(ctx context.Context) bool {
	if a.isQuitting {
		return false
	}

	runtime.WindowHide(ctx)
	return true
}

func (a *App) shutdown(ctx context.Context) {
	if a.tray != nil {
		a.tray.Dispose()
	}
}

func (a *App) showMainWindow() {
	if a.ctx == nil {
		return
	}

	runtime.WindowUnminimise(a.ctx)
	runtime.WindowShow(a.ctx)
	a.refreshFrontend()
}

func (a *App) quitFromTray() {
	if a.ctx == nil {
		return
	}

	a.isQuitting = true
	runtime.Quit(a.ctx)
}

func (a *App) refreshFrontend() {
	if a.ctx == nil {
		return
	}

	runtime.EventsEmit(a.ctx, "tray:refresh")
}

// IsNvmAvailable checks if nvm is available in the system
func (a *App) IsNvmAvailable() (bool, error) {
	return a.nvmService.CheckNvmAvailable()
}

// GetVersionList returns all installed Node.js versions
func (a *App) GetVersionList() ([]NodeVersion, error) {
	return a.nvmService.ListInstalled()
}

// GetCurrentVersion returns current environment information
func (a *App) GetCurrentVersion() (*CurrentInfo, error) {
	return a.nvmService.GetCurrent()
}

// InstallVersion installs a specific Node.js version
func (a *App) InstallVersion(version string) error {
	return a.nvmService.Install(version)
}

// UseVersion switches to a specific Node.js version
func (a *App) UseVersion(version string) error {
	return a.nvmService.Use(version)
}

// UninstallVersion removes a specific Node.js version
func (a *App) UninstallVersion(version string) error {
	return a.nvmService.Uninstall(version)
}

// GetAvailableVersions returns all available Node.js versions from remote
func (a *App) GetAvailableVersions() ([]RemoteVersion, error) {
	return a.nvmService.ListAvailable()
}

// GetGlobalNpmPackages returns globally installed npm packages for current Node environment
func (a *App) GetGlobalNpmPackages() ([]GlobalNpmPackage, error) {
	return a.nvmService.ListGlobalNpmPackages()
}

// UninstallGlobalNpmPackage removes a globally installed npm package
func (a *App) UninstallGlobalNpmPackage(name string) error {
	return a.nvmService.UninstallGlobalNpmPackage(name)
}
