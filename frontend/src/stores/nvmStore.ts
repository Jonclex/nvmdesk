import { create } from 'zustand';
import {
  GetAvailableVersions,
  GetCurrentVersion,
  GetGlobalNpmPackages,
  GetVersionList,
  InstallVersion,
  IsNvmAvailable,
  UninstallVersion,
  UseVersion,
} from '../../wailsjs/go/main/App';

export interface NodeVersion {
  version: string;
  isCurrent: boolean;
}

export interface RemoteVersion {
  version: string;
  isLTS: boolean;
}

export interface CurrentInfo {
  nodeVersion: string;
  npmVersion: string;
  nvmVersion: string;
  nvmRoot: string;
}

export interface GlobalNpmPackage {
  name: string;
  version: string;
  path: string;
  sizeBytes: number;
  sizeLabel: string;
}

export interface LogEntry {
  id: string;
  time: string;
  message: string;
  level: 'info' | 'success' | 'error' | 'warning';
}

interface NvmState {
  versions: NodeVersion[];
  availableVersions: RemoteVersion[];
  globalPackages: GlobalNpmPackage[];
  currentInfo: CurrentInfo | null;
  isNvmAvailable: boolean;
  loading: boolean;
  loadingAvailable: boolean;
  loadingPackages: boolean;
  installingVersion: string | null;
  logs: LogEntry[];

  checkNvmAvailable: () => Promise<void>;
  fetchVersions: () => Promise<void>;
  fetchAvailableVersions: () => Promise<void>;
  fetchCurrentInfo: () => Promise<void>;
  fetchGlobalPackages: () => Promise<void>;
  installVersion: (version: string) => Promise<boolean>;
  useVersion: (version: string) => Promise<boolean>;
  uninstallVersion: (version: string) => Promise<boolean>;
  addLog: (message: string, level: LogEntry['level']) => void;
  clearLogs: () => void;
  refreshAll: () => Promise<void>;
}

const formatTime = (): string => {
  const now = new Date();
  return now.toLocaleTimeString('zh-CN', { hour12: false });
};

const generateId = (): string => Date.now().toString(36) + Math.random().toString(36).slice(2);

export const useNvmStore = create<NvmState>((set, get) => ({
  versions: [],
  availableVersions: [],
  globalPackages: [],
  currentInfo: null,
  isNvmAvailable: false,
  loading: false,
  loadingAvailable: false,
  loadingPackages: false,
  installingVersion: null,
  logs: [],

  addLog: (message, level) => {
    const log: LogEntry = {
      id: generateId(),
      time: formatTime(),
      message,
      level,
    };

    set((state) => ({
      logs: [log, ...state.logs].slice(0, 200),
    }));
  },

  clearLogs: () => set({ logs: [] }),

  checkNvmAvailable: async () => {
    try {
      const available = await IsNvmAvailable();
      set({ isNvmAvailable: available });
      if (!available) {
        get().addLog('未检测到 nvm，请先安装 nvm-windows。', 'error');
      }
    } catch (error) {
      set({ isNvmAvailable: false });
      get().addLog(`检测 nvm 失败: ${error}`, 'error');
    }
  },

  fetchVersions: async () => {
    set({ loading: true });
    try {
      const versions = await GetVersionList();
      set({ versions: versions || [], loading: false });
      get().addLog('已获取本地 Node.js 版本列表。', 'info');
    } catch (error) {
      set({ versions: [], loading: false });
      get().addLog(`获取版本列表失败: ${error}`, 'error');
    }
  },

  fetchAvailableVersions: async () => {
    set({ loadingAvailable: true });
    try {
      const versions = await GetAvailableVersions();
      set({ availableVersions: versions || [], loadingAvailable: false });
      get().addLog('已获取可安装的 Node.js 版本。', 'info');
    } catch (error) {
      set({ availableVersions: [], loadingAvailable: false });
      get().addLog(`获取可安装版本失败: ${error}`, 'error');
    }
  },

  fetchCurrentInfo: async () => {
    try {
      const info = await GetCurrentVersion();
      set({ currentInfo: info });
    } catch (error) {
      get().addLog(`获取当前环境信息失败: ${error}`, 'error');
    }
  },

  fetchGlobalPackages: async () => {
    set({ loadingPackages: true });
    try {
      const packages = await GetGlobalNpmPackages();
      set({ globalPackages: packages || [], loadingPackages: false });
      get().addLog('已获取当前 Node 环境的全局 npm 包。', 'info');
    } catch (error) {
      set({ globalPackages: [], loadingPackages: false });
      get().addLog(`获取全局 npm 包失败: ${error}`, 'warning');
    }
  },

  installVersion: async (version) => {
    set({ installingVersion: version });
    get().addLog(`开始安装 Node.js ${version}...`, 'info');
    try {
      await InstallVersion(version);
      get().addLog(`Node.js ${version} 安装成功。`, 'success');
      await get().fetchVersions();
      await get().fetchCurrentInfo();
      await get().fetchGlobalPackages();
      set({ installingVersion: null });
      return true;
    } catch (error) {
      get().addLog(`安装失败: ${error}`, 'error');
      set({ installingVersion: null });
      return false;
    }
  },

  useVersion: async (version) => {
    set({ loading: true });
    get().addLog(`正在切换到 Node.js ${version}...`, 'info');
    try {
      await UseVersion(version);
      get().addLog(`已切换到 Node.js ${version}。`, 'success');
      await get().fetchVersions();
      await get().fetchCurrentInfo();
      await get().fetchGlobalPackages();
      set({ loading: false });
      return true;
    } catch (error) {
      get().addLog(`切换失败: ${error}`, 'error');
      set({ loading: false });
      return false;
    }
  },

  uninstallVersion: async (version) => {
    set({ loading: true });
    get().addLog(`正在卸载 Node.js ${version}...`, 'info');
    try {
      await UninstallVersion(version);
      get().addLog(`Node.js ${version} 已卸载。`, 'success');
      await get().fetchVersions();
      await get().fetchCurrentInfo();
      await get().fetchGlobalPackages();
      set({ loading: false });
      return true;
    } catch (error) {
      get().addLog(`卸载失败: ${error}`, 'error');
      set({ loading: false });
      return false;
    }
  },

  refreshAll: async () => {
    set({ loading: true });
    await get().checkNvmAvailable();
    if (get().isNvmAvailable) {
      await get().fetchVersions();
      await get().fetchCurrentInfo();
      await get().fetchGlobalPackages();
    }
    set({ loading: false });
  },
}));
