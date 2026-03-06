package main

import (
	"context"
)

// App struct - main application structure exposed to frontend
type App struct {
	ctx        context.Context
	nvmService *NvmService
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
