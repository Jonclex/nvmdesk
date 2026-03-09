package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// NvmService handles all nvm command operations
type NvmService struct{}

type npmPackageMeta struct {
	Version string `json:"version"`
}

type npmListOutput struct {
	Dependencies map[string]npmPackageMeta `json:"dependencies"`
}

// NewNvmService creates a new NvmService instance
func NewNvmService() *NvmService {
	return &NvmService{}
}

// Version number regex pattern
var versionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// ValidateVersion checks if the version string is valid
func (s *NvmService) ValidateVersion(version string) bool {
	return versionPattern.MatchString(version)
}

// execCommand executes an nvm command with timeout
func (s *NvmService) execCommand(timeout time.Duration, args ...string) (string, error) {
	return s.execNamedCommand(timeout, "nvm", args...)
}

// execNamedCommand executes a command with timeout and hidden window
func (s *NvmService) execNamedCommand(timeout time.Duration, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = os.Environ()
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000,
	}

	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("命令执行超时")
	}
	if err != nil {
		return string(output), fmt.Errorf("%s: %s", err.Error(), string(output))
	}
	return string(output), nil
}

// CheckNvmAvailable checks if nvm is available in the system
func (s *NvmService) CheckNvmAvailable() (bool, error) {
	_, err := exec.LookPath("nvm")
	if err != nil {
		return false, nil
	}

	output, err := s.execCommand(10*time.Second, "version")
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(output) != "", nil
}

// GetNvmVersion returns the nvm version
func (s *NvmService) GetNvmVersion() (string, error) {
	output, err := s.execCommand(10*time.Second, "version")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// GetNvmRoot returns the nvm root directory
func (s *NvmService) GetNvmRoot() (string, error) {
	output, err := s.execCommand(10*time.Second, "root")
	if err != nil {
		return "", err
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "Current") {
			return line, nil
		}
	}
	return strings.TrimSpace(output), nil
}

// ListInstalled returns all installed Node.js versions
func (s *NvmService) ListInstalled() ([]NodeVersion, error) {
	output, err := s.execCommand(30*time.Second, "list")
	if err != nil {
		return nil, err
	}

	var versions []NodeVersion
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "No installations") {
			continue
		}

		isCurrent := strings.HasPrefix(line, "*")
		line = strings.TrimPrefix(line, "*")
		line = strings.TrimSpace(line)

		parts := strings.Fields(line)
		if len(parts) > 0 {
			version := parts[0]
			if versionPattern.MatchString(version) {
				versions = append(versions, NodeVersion{
					Version:   version,
					IsCurrent: isCurrent,
				})
			}
		}
	}

	return versions, nil
}

// execHiddenCommand executes a command with hidden window
func (s *NvmService) execHiddenCommand(name string, args ...string) (string, error) {
	return s.execNamedCommand(10*time.Second, name, args...)
}

// GetCurrent returns the current Node.js and npm versions
func (s *NvmService) GetCurrent() (*CurrentInfo, error) {
	info := &CurrentInfo{}

	nvmVersion, err := s.GetNvmVersion()
	if err == nil {
		info.NvmVersion = nvmVersion
	}

	nvmRoot, err := s.GetNvmRoot()
	if err == nil {
		info.NvmRoot = nvmRoot
	}

	currentOutput, err := s.execCommand(10*time.Second, "current")
	if err == nil {
		info.NodeVersion = strings.TrimSpace(currentOutput)
	}

	if nodeOutput, err := s.execHiddenCommand("node", "-v"); err == nil {
		info.NodeVersion = strings.TrimPrefix(strings.TrimSpace(nodeOutput), "v")
	}

	if npmOutput, err := s.execHiddenCommand("npm", "-v"); err == nil {
		info.NpmVersion = strings.TrimSpace(npmOutput)
	}

	return info, nil
}

// Install installs a specific Node.js version
func (s *NvmService) Install(version string) error {
	if !s.ValidateVersion(version) {
		return fmt.Errorf(ErrorMessages[ErrVersionInvalid])
	}

	_, err := s.execCommand(10*time.Minute, "install", version)
	if err != nil {
		if strings.Contains(err.Error(), "access") || strings.Contains(err.Error(), "permission") {
			return fmt.Errorf(ErrorMessages[ErrPermissionDenied])
		}
		if strings.Contains(err.Error(), "network") || strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "connection") {
			return fmt.Errorf(ErrorMessages[ErrNetworkError])
		}
		return fmt.Errorf("安装失败: %s", err.Error())
	}
	return nil
}

// Use switches to a specific Node.js version
func (s *NvmService) Use(version string) error {
	if !s.ValidateVersion(version) {
		return fmt.Errorf(ErrorMessages[ErrVersionInvalid])
	}

	_, err := s.execCommand(30*time.Second, "use", version)
	if err != nil {
		if strings.Contains(err.Error(), "not installed") {
			return fmt.Errorf(ErrorMessages[ErrVersionNotInstalled])
		}
		if strings.Contains(err.Error(), "access") || strings.Contains(err.Error(), "permission") {
			return fmt.Errorf(ErrorMessages[ErrPermissionDenied])
		}
		return fmt.Errorf("切换失败: %s", err.Error())
	}
	return nil
}

// Uninstall removes a specific Node.js version
func (s *NvmService) Uninstall(version string) error {
	if !s.ValidateVersion(version) {
		return fmt.Errorf(ErrorMessages[ErrVersionInvalid])
	}

	_, err := s.execCommand(60*time.Second, "uninstall", version)
	if err != nil {
		if strings.Contains(err.Error(), "access") || strings.Contains(err.Error(), "permission") {
			return fmt.Errorf(ErrorMessages[ErrPermissionDenied])
		}
		return fmt.Errorf("卸载失败: %s", err.Error())
	}
	return nil
}

// parseVersion converts version string to comparable integers
func parseVersion(version string) (major, minor, patch int) {
	parts := strings.Split(version, ".")
	if len(parts) >= 1 {
		major, _ = strconv.Atoi(parts[0])
	}
	if len(parts) >= 2 {
		minor, _ = strconv.Atoi(parts[1])
	}
	if len(parts) >= 3 {
		patch, _ = strconv.Atoi(parts[2])
	}
	return
}

// compareVersions compares two version strings, returns true if v1 > v2
func compareVersions(v1, v2 string) bool {
	major1, minor1, patch1 := parseVersion(v1)
	major2, minor2, patch2 := parseVersion(v2)

	if major1 != major2 {
		return major1 > major2
	}
	if minor1 != minor2 {
		return minor1 > minor2
	}
	return patch1 > patch2
}

// ListAvailable returns all available Node.js versions from remote
func (s *NvmService) ListAvailable() ([]RemoteVersion, error) {
	output, err := s.execCommand(60*time.Second, "list", "available")
	if err != nil {
		return nil, fmt.Errorf("获取可用版本失败: %s", err.Error())
	}

	var versions []RemoteVersion
	versionSet := make(map[string]bool)
	lines := strings.Split(output, "\n")

	headerPassed := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.Contains(line, "---") || strings.Contains(line, "LTS") || strings.Contains(line, "CURRENT") {
			headerPassed = true
			continue
		}

		if !headerPassed {
			continue
		}

		parts := strings.Fields(line)
		for _, part := range parts {
			part = strings.TrimSpace(part)
			part = strings.Trim(part, "|")
			part = strings.TrimSpace(part)
			if versionPattern.MatchString(part) && !versionSet[part] {
				versionSet[part] = true
				versions = append(versions, RemoteVersion{
					Version: part,
					IsLTS:   false,
				})
			}
		}
	}

	sort.Slice(versions, func(i, j int) bool {
		return compareVersions(versions[i].Version, versions[j].Version)
	})

	if len(versions) > 20 {
		versions = versions[:20]
	}

	return versions, nil
}

// ListGlobalNpmPackages returns globally installed npm packages for the current Node environment
func (s *NvmService) ListGlobalNpmPackages() ([]GlobalNpmPackage, error) {
	if _, err := exec.LookPath("npm"); err != nil {
		return nil, fmt.Errorf("未检测到 npm，请先切换到可用的 Node.js 版本")
	}

	rootOutput, err := s.execNamedCommand(15*time.Second, "npm", "root", "-g")
	if err != nil {
		return nil, fmt.Errorf("获取全局 npm 目录失败: %s", err.Error())
	}

	globalRoot := strings.TrimSpace(rootOutput)
	if globalRoot == "" {
		return []GlobalNpmPackage{}, nil
	}

	listOutput, err := s.execNamedCommand(30*time.Second, "npm", "ls", "-g", "--depth=0", "--json")
	if err != nil {
		return nil, fmt.Errorf("获取全局 npm 包列表失败: %s", err.Error())
	}

	var npmList npmListOutput
	if err := json.Unmarshal([]byte(listOutput), &npmList); err != nil {
		return nil, fmt.Errorf("解析全局 npm 包列表失败: %s", err.Error())
	}

	packages := make([]GlobalNpmPackage, 0, len(npmList.Dependencies))
	for name, meta := range npmList.Dependencies {
		packagePath := filepath.Join(globalRoot, filepath.FromSlash(name))
		sizeBytes, sizeErr := calculateDirectorySize(packagePath)
		if sizeErr != nil && !os.IsNotExist(sizeErr) {
			sizeBytes = 0
		}

		packages = append(packages, GlobalNpmPackage{
			Name:      name,
			Version:   meta.Version,
			Path:      packagePath,
			SizeBytes: sizeBytes,
			SizeLabel: formatBytes(sizeBytes),
		})
	}

	sort.Slice(packages, func(i, j int) bool {
		if packages[i].SizeBytes == packages[j].SizeBytes {
			return packages[i].Name < packages[j].Name
		}
		return packages[i].SizeBytes > packages[j].SizeBytes
	})

	return packages, nil
}

func calculateDirectorySize(root string) (int64, error) {
	var totalSize int64
	err := filepath.WalkDir(root, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		totalSize += info.Size()
		return nil
	})
	return totalSize, err
}

func formatBytes(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.2f %ciB", float64(size)/float64(div), "KMGTPE"[exp])
}
