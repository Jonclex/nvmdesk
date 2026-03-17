# nvm 桌面管理实现文档

## 1. 文档目的

本文档描述 `nvmdesk` 当前已落地的实现方案，重点说明：

- 当前技术栈与模块划分
- 已实现功能与行为边界
- Windows 托盘集成方式
- 现阶段未覆盖的能力

本文档以仓库当前代码为准，不再使用早期的 Tauri 方案设计稿描述。

## 2. 当前实现概览

`nvmdesk` 是一个基于 `Wails v2 + React + Ant Design` 的 Windows 桌面应用，用于管理本机 `nvm-windows` 环境下的 Node.js 版本。

当前支持的核心能力：

- 检测 `nvm` 是否可用
- 获取已安装 Node.js 版本列表
- 安装指定 Node.js 版本
- 切换当前 Node.js 版本
- 卸载已安装 Node.js 版本
- 获取当前 `Node.js / npm / nvm / nvm root` 信息
- 查看当前 Node 环境下的全局 npm 包
- 卸载当前 Node 环境下的全局 npm 包
- 记录前端操作日志
- Windows 最小化到托盘，并通过托盘菜单切换版本/刷新/退出

当前实现目标平台：

- 已实现：Windows
- 兼容占位：非 Windows 平台通过 `tray_stub.go` 提供空托盘实现
- 未实现：macOS/Linux 的 `nvm-sh` 适配

## 3. 技术栈

### 3.1 后端

- Go
- Wails v2
- `os/exec` 调用 `nvm`、`node`、`npm`
- `golang.org/x/sys/windows` 实现 Windows 托盘能力

### 3.2 前端

- React 18
- TypeScript
- Vite
- Ant Design 5
- Zustand

## 4. 架构划分

当前项目采用前后端分层，但整体保持轻量：

### 4.1 Wails 应用层

入口文件：

- `main.go`
- `app.go`

职责：

- 初始化 Wails 窗口
- 绑定 Go 方法给前端调用
- 处理启动、关闭、托盘退出等生命周期
- 通过 Wails 事件向前端发送刷新通知

### 4.2 NVM 服务层

核心文件：

- `nvm_service.go`
- `types.go`

职责：

- 校验版本号格式
- 统一封装白名单命令调用
- 映射用户可读错误信息
- 聚合当前环境信息
- 解析远程版本与全局 npm 包数据

### 4.3 前端展示层

核心文件：

- `frontend/src/App.tsx`
- `frontend/src/stores/nvmStore.ts`
- `frontend/src/components/*`

职责：

- 展示状态卡片、版本列表、安装弹窗、npm 包列表、日志面板
- 管理加载状态、日志、安装中状态、删除中状态
- 调用 Wails 暴露的方法并刷新界面

### 4.4 Windows 托盘层

核心文件：

- `tray_windows.go`
- `tray_stub.go`

职责：

- 劫持窗口消息
- 在最小化时隐藏窗口到系统托盘
- 响应托盘左键恢复主窗口
- 响应托盘右键菜单
- 通过托盘菜单切换 Node.js 版本

## 5. 核心模块说明

## 5.1 App

`App` 是前端调用入口，也是窗口与托盘的协调中心。

主要公开方法：

- `IsNvmAvailable()`
- `GetVersionList()`
- `GetCurrentVersion()`
- `InstallVersion(version string)`
- `UseVersion(version string)`
- `UninstallVersion(version string)`
- `GetAvailableVersions()`
- `GetGlobalNpmPackages()`
- `UninstallGlobalNpmPackage(name string)`

内部辅助行为：

- `startup()` 中初始化托盘
- `beforeClose()` 中拦截窗口关闭并隐藏窗口
- `quitFromTray()` 中标记真正退出
- `refreshFrontend()` 通过 `tray:refresh` 事件通知前端刷新

## 5.2 NvmService

`NvmService` 负责所有命令执行与解析逻辑。

### 已实现方法

- `CheckNvmAvailable()`
- `GetNvmVersion()`
- `GetNvmRoot()`
- `ListInstalled()`
- `GetCurrent()`
- `Install(version string)`
- `Use(version string)`
- `Uninstall(version string)`
- `ListAvailable()`
- `ListGlobalNpmPackages()`
- `UninstallGlobalNpmPackage(name string)`

### 命令执行策略

- 统一使用 `exec.CommandContext`
- 所有命令带超时控制
- Windows 下隐藏子进程窗口
- 只允许调用固定命令：`nvm`、`node`、`npm`
- 版本号必须匹配正则：`^\d+\.\d+\.\d+$`

### 当前超时设置

- `nvm version` / `nvm root` / `nvm current`：10 秒
- `nvm list`：30 秒
- `nvm install`：10 分钟
- `nvm uninstall`：60 秒
- `nvm list available`：60 秒
- `npm uninstall -g`：2 分钟

## 5.3 前端状态管理

`frontend/src/stores/nvmStore.ts` 使用 Zustand 管理全局状态。

主要状态：

- `versions`
- `availableVersions`
- `globalPackages`
- `currentInfo`
- `isNvmAvailable`
- `loading`
- `loadingAvailable`
- `loadingPackages`
- `installingVersion`
- `deletingPackage`
- `logs`

主要行为：

- `refreshAll()`
- `fetchVersions()`
- `fetchCurrentInfo()`
- `fetchAvailableVersions()`
- `fetchGlobalPackages()`
- `installVersion()`
- `useVersion()`
- `uninstallVersion()`
- `uninstallGlobalPackage()`
- `addLog()`
- `clearLogs()`

日志保存在前端内存中，最多保留最近 200 条，不落盘。

## 5.4 Windows 托盘实现

托盘能力仅在 Windows 编译目标下启用。

### 触发逻辑

- 窗口最小化时，拦截 `wmSize`
- 当 `wParam == sizeMinimized` 时隐藏主窗口
- 托盘左键单击或双击时恢复主窗口
- 托盘右键弹出上下文菜单

### 托盘菜单内容

- 显示主窗口
- 切换 Node.js
- 刷新版本列表
- 退出

其中“切换 Node.js”子菜单会：

- 拉取当前已安装版本列表
- 按版本号从高到低排序
- 对当前生效版本打勾
- 点击后异步调用 `nvm use`

### 托盘通知

以下操作会通过气泡通知反馈结果：

- 托盘切换 Node.js 成功/失败
- 托盘刷新版本列表成功/失败

成功操作后会向前端广播 `tray:refresh`，触发界面重新拉取数据。

## 6. 页面与组件

当前主界面由以下模块组成：

### 6.1 当前环境卡片

文件：

- `frontend/src/components/StatusCard.tsx`

展示内容：

- 当前 Node.js 版本
- 当前 npm 版本
- 当前 nvm 版本
- nvm 根目录

如果未检测到 `nvm`，显示错误提示卡片。

### 6.2 已安装版本列表

文件：

- `frontend/src/components/VersionList.tsx`
- `frontend/src/components/InstallModal.tsx`

提供功能：

- 刷新版本列表
- 打开安装弹窗
- 切换某个版本
- 卸载某个非当前版本
- 查看当前版本标记

安装弹窗支持两种方式：

- 从远程可用版本列表中选择安装
- 手动输入版本号安装

### 6.3 全局 npm 包列表

文件：

- `frontend/src/components/NpmPackageList.tsx`

提供功能：

- 展示当前 Node 环境下的全局 npm 包
- 显示包版本、安装路径、体积
- 按体积排序
- 卸载指定全局 npm 包

### 6.4 操作日志面板

文件：

- `frontend/src/components/LogPanel.tsx`

提供功能：

- 展示前端操作日志
- 支持清空日志
- 显示日志级别：`info / success / warning / error`

## 7. 数据结构

后端核心结构定义位于 `types.go`。

### 7.1 已安装版本

```go
type NodeVersion struct {
    Version   string `json:"version"`
    IsCurrent bool   `json:"isCurrent"`
}
```

### 7.2 远程可安装版本

```go
type RemoteVersion struct {
    Version string `json:"version"`
    IsLTS   bool   `json:"isLTS"`
}
```

说明：

- 当前 `ListAvailable()` 已返回 `IsLTS` 字段
- 但现阶段解析逻辑未真正区分 LTS 与 Current
- 前端当前也未展示 LTS 标识

### 7.3 当前环境信息

```go
type CurrentInfo struct {
    NodeVersion string `json:"nodeVersion"`
    NpmVersion  string `json:"npmVersion"`
    NvmVersion  string `json:"nvmVersion"`
    NvmRoot     string `json:"nvmRoot"`
}
```

### 7.4 全局 npm 包

```go
type GlobalNpmPackage struct {
    Name      string `json:"name"`
    Version   string `json:"version"`
    Path      string `json:"path"`
    SizeBytes int64  `json:"sizeBytes"`
    SizeLabel string `json:"sizeLabel"`
}
```

## 8. 错误处理

当前定义的错误码位于 `types.go`：

- `NVM_NOT_FOUND`
- `VERSION_INVALID`
- `VERSION_NOT_INSTALLED`
- `NETWORK_ERROR`
- `PERMISSION_DENIED`
- `COMMAND_FAILED`

当前实际映射策略：

- 版本号非法时返回 `VERSION_INVALID`
- `nvm use` 命中 `not installed` 时返回 `VERSION_NOT_INSTALLED`
- 安装过程命中网络相关关键字时返回 `NETWORK_ERROR`
- 命中权限相关关键字时返回 `PERMISSION_DENIED`

需要注意：

- 当前多数错误最终仍以中文字符串直接返回
- 错误码结构体 `AppError` 已定义，但还未在前后端之间统一使用

## 9. 当前项目结构

```text
nvmdesk/
├── app.go
├── main.go
├── nvm_service.go
├── tray_windows.go
├── tray_stub.go
├── types.go
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── InstallModal.tsx
│   │   │   ├── LogPanel.tsx
│   │   │   ├── NpmPackageList.tsx
│   │   │   ├── StatusCard.tsx
│   │   │   └── VersionList.tsx
│   │   ├── stores/
│   │   │   └── nvmStore.ts
│   │   ├── App.tsx
│   │   └── main.tsx
│   └── package.json
└── docs/
    └── nvm桌面管理实现文档.md
```

## 10. 已实现与未实现边界

### 10.1 已实现

- Windows 桌面应用基础框架
- `nvm-windows` 可用性检测
- Node.js 版本安装、切换、卸载
- 当前环境信息展示
- 远程可安装版本拉取
- 当前 Node 环境的全局 npm 包查看与卸载
- 前端操作日志
- 最小化到托盘与托盘版本切换

### 10.2 未实现

- `.nvmrc` 项目识别与一键切换
- 镜像源配置管理
- npm registry 配置
- 日志落盘与导出
- 安装过程实时流式日志
- macOS/Linux 适配
- 自动启动 / 开机自启
- 更细粒度的结构化错误码协议

## 11. 后续建议

建议下一阶段优先处理以下事项：

1. 补齐 `.nvmrc` 项目切换能力，完成“按项目使用 Node 版本”的核心闭环。
2. 统一错误返回格式，真正使用 `AppError` 进行前后端协议化传输。
3. 将日志从前端内存扩展为本地持久化，便于排障。
4. 为 `ListAvailable()` 增加准确的 LTS 标记解析。
5. 为托盘能力补充自动启动、启动即隐藏等桌面端行为配置。
