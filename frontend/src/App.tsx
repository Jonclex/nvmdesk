import { useEffect } from 'react';
import { ConfigProvider, Layout, Typography, Spin } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import StatusCard from './components/StatusCard';
import VersionList from './components/VersionList';
import LogPanel from './components/LogPanel';
import { useNvmStore } from './stores/nvmStore';

const { Header, Content } = Layout;
const { Title } = Typography;

function App() {
  const { refreshAll, loading, isNvmAvailable } = useNvmStore();

  useEffect(() => {
    refreshAll();
  }, []);

  return (
    <ConfigProvider locale={zhCN}>
      <Layout style={{ minHeight: '100vh', background: '#f5f5f5' }}>
        <Header
          style={{
            background: '#fff',
            padding: '0 24px',
            display: 'flex',
            alignItems: 'center',
            borderBottom: '1px solid #f0f0f0',
            height: 56,
          }}
        >
          <Title level={4} style={{ margin: 0, color: '#1890ff' }}>
            NVM Desktop Manager
          </Title>
        </Header>
        <Content style={{ padding: 24 }}>
          <Spin spinning={loading && !isNvmAvailable} tip="正在检测 NVM...">
            <StatusCard />
            <VersionList />
            <LogPanel />
          </Spin>
        </Content>
      </Layout>
    </ConfigProvider>
  );
}

export default App;
