# nvm 桌面管理实现文档

## 1. 目标与范围
- 目标：实现一个可视化桌面应用，用于管理本机 Node.js 版本（基于 nvm）。
- 适用用户：前端开发者、全栈开发者、需要在多个项目间切换 Node 版本的团队成员。
- 平台范围：Windows 优先（nvm-windows），后续支持 macOS/Linux（nvm-sh）。

## 2. 核心功能
1. Node 版本管理
- 获取已安装版本列表。
- 安装指定版本（如 `18.20.8`、`20.19.0`）。
- 切换当前版本（`nvm use`）。
- 卸载指定版本。

2. 环境状态展示
- 展示当前生效 Node/npm 版本。
- 展示 nvm 根目录、Node 安装目录、PATH 状态。
- 检测 nvm 可执行文件是否可用。

3. 项目级版本切换
- 读取项目 `.nvmrc`。
- 一键切换到项目要求版本。
- 可选：进入项目目录时自动提示切换。

4. 镜像与网络配置
- 设置 Node 下载镜像（官方、淘宝镜像、公司内网镜像）。
- 设置 npm registry。
- 安装失败时输出可读错误信息与排查建议。

5. 操作历史与日志
- 记录关键操作（安装、切换、卸载）及结果。
- 支持导出日志用于排障。

## 3. 非功能需求
- 可用性：关键操作 3 步内可完成。
- 稳定性：命令执行失败可回滚 UI 状态，不出现假成功。
- 安全性：不执行用户任意命令，仅允许白名单命令。
- 性能：常规命令结果 2 秒内反馈；安装过程实时流式日志。

## 4. 技术方案
## 4.1 推荐技术栈
- 桌面框架：Tauri（Rust 后端 + Web 前端）。
- 前端：React + TypeScript + Vite。
- UI：Ant Design 或 Mantine（快速构建表单/表格/通知）。
- 状态管理：Zustand。
- 日志：本地文件 + 前端操作记录面板。

> 说明：如果团队更熟悉 Electron，可替换为 Electron + Node 主进程，核心模块边界不变。

## 4.2 架构分层
1. UI 层（前端）
- 版本列表、安装表单、状态卡片、日志面板。

2. 应用服务层
- 统一封装 `nvm` 操作接口。
- 处理错误映射（命令错误 -> 用户可读文案）。

3. 系统适配层
- Windows: 调用 `nvm.exe`。
- macOS/Linux: 调用 shell `nvm`（需加载 shell profile）。

4. 持久化层
- 本地配置文件（镜像地址、默认版本、最近项目）。
- 本地日志文件。

## 5. 模块设计
## 5.1 nvm 命令网关（NvmGateway）
- `listInstalled(): Promise<string[]>`
- `listRemote(): Promise<string[]>`（可选）
- `install(version: string): Promise<void>`
- `use(version: string): Promise<void>`
- `uninstall(version: string): Promise<void>`
- `current(): Promise<{ node: string; npm: string; nvm: string }>`
- `setMirror(nodeMirror: string): Promise<void>`

实现要点：
- 只允许调用白名单命令，禁止拼接任意 shell 输入。
- 对版本号做正则校验：`^\d+\.\d+\.\d+$`。
- 命令执行使用超时控制（如 10 分钟）。

## 5.2 项目扫描模块（ProjectResolver）
- 读取指定路径下 `.nvmrc`。
- 校验版本格式。
- 若未安装则提示“先安装后切换”。

## 5.3 配置管理模块（SettingsService）
- 保存：
  - 默认镜像地址
  - 是否自动检测 `.nvmrc`
  - 最近项目路径
- 配置文件建议路径：
  - Windows: `%APPDATA%/nvmdesk/config.json`
  - macOS/Linux: `~/.config/nvmdesk/config.json`

## 5.4 日志模块（LogService）
- 日志级别：INFO/WARN/ERROR。
- 日志落盘 + UI 展示最近 200 条。
- 异常堆栈只在调试模式展示详情。

## 6. 关键流程
## 6.1 安装版本流程
1. 输入版本号 -> 前端校验格式。
2. 调用 `NvmGateway.install(version)`。
3. 实时显示命令输出。
4. 成功后刷新版本列表与当前版本信息。
5. 失败时展示错误原因与建议（镜像、网络、权限）。

## 6.2 切换版本流程
1. 用户点击“切换”。
2. 调用 `NvmGateway.use(version)`。
3. 调用 `current()` 二次校验切换结果。
4. 更新 UI，并记录日志。

## 6.3 项目一键切换流程
1. 用户选择项目目录。
2. 读取 `.nvmrc`。
3. 检查目标版本是否已安装。
4. 已安装：直接 `use`；未安装：弹窗确认安装后切换。

## 7. 错误码与提示规范
- `NVM_NOT_FOUND`: 未检测到 nvm，请先安装 nvm 并重启应用。
- `VERSION_INVALID`: 版本号格式错误，请输入 `x.y.z`。
- `VERSION_NOT_INSTALLED`: 目标版本未安装，请先安装。
- `NETWORK_ERROR`: 下载失败，请检查网络或切换镜像。
- `PERMISSION_DENIED`: 权限不足，请以管理员身份重试（Windows）。

## 8. 目录建议
```text
nvmdesk/
  src/
    ui/
    services/
      nvm-gateway.ts
      settings-service.ts
      log-service.ts
      project-resolver.ts
    types/
  src-tauri/              # 若使用 Tauri
  docs/
    nvm桌面管理实现文档.md
```

## 9. 里程碑计划
1. M1（1 周）：完成 UI 原型 + nvm 基础命令联调（list/install/use）。
2. M2（1 周）：完成项目 `.nvmrc` 一键切换、配置页、日志面板。
3. M3（3-5 天）：异常处理、安装体验优化、打包发布（Windows）。
4. M4（可选）：跨平台支持（macOS/Linux）和自动检测增强。

## 10. 验收标准
- 可成功安装、切换、卸载至少 3 个 Node 版本。
- 切换后 `node -v` 与 UI 显示一致。
- `.nvmrc` 项目可一键切换，失败有明确提示。
- 离线/弱网场景下错误可定位、可重试。
- 日志可导出并包含关键上下文（时间、命令、结果）。

## 11. 后续增强
- 支持多源镜像测速并自动推荐最快源。
- 支持团队策略（项目最低 Node 版本、白名单版本）。
- 支持检测并修复 PATH 异常。
- 支持导入导出团队统一配置。
