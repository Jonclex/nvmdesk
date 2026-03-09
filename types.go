package main

// NodeVersion represents an installed Node.js version
type NodeVersion struct {
	Version   string `json:"version"`
	IsCurrent bool   `json:"isCurrent"`
}

// RemoteVersion represents an available Node.js version from remote
type RemoteVersion struct {
	Version string `json:"version"`
	IsLTS   bool   `json:"isLTS"`
}

// CurrentInfo represents current environment information
type CurrentInfo struct {
	NodeVersion string `json:"nodeVersion"`
	NpmVersion  string `json:"npmVersion"`
	NvmVersion  string `json:"nvmVersion"`
	NvmRoot     string `json:"nvmRoot"`
}

// GlobalNpmPackage represents a globally installed npm package
type GlobalNpmPackage struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Path      string `json:"path"`
	SizeBytes int64  `json:"sizeBytes"`
	SizeLabel string `json:"sizeLabel"`
}

// LogEntry represents a log entry
type LogEntry struct {
	Time    string `json:"time"`
	Message string `json:"message"`
	Level   string `json:"level"`
}

// AppError represents application error codes
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error codes
const (
	ErrNvmNotFound        = "NVM_NOT_FOUND"
	ErrVersionInvalid     = "VERSION_INVALID"
	ErrVersionNotInstalled = "VERSION_NOT_INSTALLED"
	ErrNetworkError       = "NETWORK_ERROR"
	ErrPermissionDenied   = "PERMISSION_DENIED"
	ErrCommandFailed      = "COMMAND_FAILED"
)

// Error messages mapping
var ErrorMessages = map[string]string{
	ErrNvmNotFound:        "未检测到 nvm，请先安装 nvm-windows 并重启应用",
	ErrVersionInvalid:     "版本号格式错误，请输入 x.y.z 格式",
	ErrVersionNotInstalled: "该版本未安装，请先安装",
	ErrNetworkError:       "下载失败，请检查网络或更换镜像",
	ErrPermissionDenied:   "权限不足，请以管理员身份运行",
	ErrCommandFailed:      "命令执行失败",
}
