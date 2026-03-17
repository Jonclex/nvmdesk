# NVM Desktop Manager

一个基于 `Wails + Go + React + TypeScript` 的 Windows 桌面应用，用于可视化管理本机 `nvm-windows` 环境下的 Node.js 版本。

![screenshot](docs/screenshot.png)

## 功能特性

- 检测本机 `nvm` 是否可用
- 查看已安装的 Node.js 版本列表
- 安装新的 Node.js 版本
- 切换当前 Node.js 版本
- 卸载已安装的 Node.js 版本
- 查看当前 `Node.js / npm / nvm / nvm root` 信息
- 查看当前 Node 环境下的全局 npm 包
- 卸载当前 Node 环境下的全局 npm 包
- 记录前端操作日志
- 支持 Windows 托盘
- 最小化后隐藏到托盘
- 通过托盘菜单快速切换 Node.js 版本、刷新版本列表、退出应用

## 当前平台支持

- 已支持：Windows
- 未支持：macOS、Linux

当前项目面向 `nvm-windows`，并未适配 `nvm-sh`。

## 技术栈

- 后端：Go + Wails v2
- 前端：React 18 + TypeScript + Vite
- UI：Ant Design
- 状态管理：Zustand
- 托盘集成：`golang.org/x/sys/windows`

## 运行前提

开始前请确保本机已满足以下条件：

- 已安装 [nvm-windows](https://github.com/coreybutler/nvm-windows)
- `nvm` 已加入系统 `PATH`
- 已安装 Go 1.23 或更高版本
- 已安装 Node.js 16+ 与 npm
- 已安装 [Wails CLI](https://wails.io/docs/gettingstarted/installation)

可先在终端确认：

```bash
nvm version
go version
wails doctor
```

如果 `nvm version` 无法执行，应用启动后会提示 `nvm` 不可用。

## 开发

```bash
git clone https://github.com/Jonclex/nvmdesk.git
cd nvmdesk

cd frontend
npm install
cd ..

wails dev
```

开发模式下：

- 前端由 Vite 提供热更新
- Go 后端通过 Wails 绑定到前端
- Windows 托盘相关行为以本机桌面环境为准

## 构建

```bash
wails build
```

构建完成后，产物位于 `build/bin/` 目录。

如果只是单独构建前端，也可以执行：

```bash
cd frontend
npm run build
```

## 使用说明

### 1. 启动应用

启动后应用会检测 `nvm` 是否可用。

- 若可用，主界面会加载当前环境信息、已安装版本、全局 npm 包
- 若不可用，界面会提示先安装 `nvm-windows` 并确认 `PATH`

### 2. 安装 Node.js 版本

“安装新版本”弹窗支持两种方式：

- 从远程版本列表中直接选择安装
- 手动输入版本号安装，例如 `20.19.0`

版本号必须为 `x.y.z` 格式。

### 3. 切换和卸载版本

在“已安装版本”列表中可以：

- 点击“切换”将当前环境切到指定版本
- 点击“卸载”删除非当前版本

切换成功后，界面会刷新当前环境信息和全局 npm 包列表。

### 4. 管理全局 npm 包

应用会展示当前 Node 环境下的全局 npm 包，并显示：

- 包名
- 版本
- 安装路径
- 目录体积

支持直接卸载全局 npm 包。

### 5. 托盘行为

Windows 下应用带有系统托盘能力：

- 点击窗口关闭按钮时，不会直接退出，而是隐藏窗口
- 窗口最小化时会隐藏到托盘
- 左键托盘图标可恢复主窗口
- 右键托盘图标可打开菜单

托盘菜单包括：

- 显示主窗口
- 切换 Node.js
- 刷新版本列表
- 退出

托盘切换版本或刷新版本列表后，会通过气泡通知反馈结果，并同步刷新主界面。

## 项目结构

```text
nvmdesk/
├── app.go
├── main.go
├── nvm_service.go
├── tray_windows.go
├── tray_stub.go
├── types.go
├── wails.json
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   ├── stores/
│   │   ├── App.tsx
│   │   └── main.tsx
│   └── package.json
├── build/
└── docs/
```

核心文件说明：

- `app.go`：Wails 应用入口与前端绑定方法
- `nvm_service.go`：`nvm`、`node`、`npm` 命令封装
- `tray_windows.go`：Windows 托盘实现
- `tray_stub.go`：非 Windows 平台托盘占位实现
- `frontend/src/stores/nvmStore.ts`：前端状态管理

## 已知限制

- 仅支持 Windows 和 `nvm-windows`
- 暂未支持 `.nvmrc` 项目识别与一键切换
- 暂未支持镜像源配置
- 暂未支持日志落盘与导出
- 暂未支持安装过程实时日志流
- 暂未支持 macOS / Linux

## 文档

- 实现说明见 [docs/nvm桌面管理实现文档.md](/f:/200_副业兼职/210_工作空间/211_产品信息/AI项目/ai-test/nvmdesk/docs/nvm桌面管理实现文档.md)

## License

MIT
