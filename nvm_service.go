package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// NvmService handles all nvm command operations
type NvmService struct{}

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
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "nvm", args...)
	cmd.Env = os.Environ()

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

	nodeCmd := exec.Command("node", "-v")
	if nodeOutput, err := nodeCmd.Output(); err == nil {
		info.NodeVersion = strings.TrimPrefix(strings.TrimSpace(string(nodeOutput)), "v")
	}

	npmCmd := exec.Command("npm", "-v")
	if npmOutput, err := npmCmd.Output(); err == nil {
		info.NpmVersion = strings.TrimSpace(string(npmOutput))
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
	
	// nvm list available 输出格式是表格形式，包含 LTS 和 Current 版本
	// 跳过表头行
	headerPassed := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// 跳过分隔线和表头
		if strings.Contains(line, "---") || strings.Contains(line, "LTS") || strings.Contains(line, "CURRENT") {
			headerPassed = true
			continue
		}
		
		if !headerPassed {
			continue
		}

		// 解析每行的版本号
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

	// 按版本号降序排序（最新版本在前）
	sort.Slice(versions, func(i, j int) bool {
		return compareVersions(versions[i].Version, versions[j].Version)
	})

	// 只返回最新的 20 个版本
	if len(versions) > 20 {
		versions = versions[:20]
	}

	return versions, nil
}
