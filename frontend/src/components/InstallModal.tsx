import React, { useState, useEffect } from 'react';
import { Modal, Input, Form, Alert, Typography, Tabs, Table, Button, Spin, Tag } from 'antd';
import { ReloadOutlined, DownloadOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { useNvmStore, RemoteVersion } from '../stores/nvmStore';

const { Text } = Typography;

interface InstallModalProps {
  open: boolean;
  onClose: () => void;
}

const InstallModal: React.FC<InstallModalProps> = ({ open, onClose }) => {
  const [form] = Form.useForm();
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<string>('select');
  const {
    installVersion,
    installingVersion,
    availableVersions,
    fetchAvailableVersions,
    loadingAvailable,
    versions,
  } = useNvmStore();

  const versionRegex = /^\d+\.\d+\.\d+$/;

  useEffect(() => {
    if (open && availableVersions.length === 0) {
      fetchAvailableVersions();
    }
  }, [open]);

  const installedSet = new Set(versions.map((v) => v.version));

  const handleInstall = async (version: string) => {
    setError(null);
    const success = await installVersion(version);
    if (success) {
      form.resetFields();
      onClose();
    }
  };

  const handleOk = async () => {
    if (activeTab === 'manual') {
      try {
        const values = await form.validateFields();
        setError(null);
        await handleInstall(values.version);
      } catch (err) {
        if (err instanceof Error) {
          setError(err.message);
        }
      }
    }
  };

  const handleCancel = () => {
    if (!installingVersion) {
      form.resetFields();
      setError(null);
      onClose();
    }
  };

  const columns: ColumnsType<RemoteVersion> = [
    {
      title: '版本',
      dataIndex: 'version',
      key: 'version',
      width: 120,
      render: (version: string) => (
        <span style={{ fontFamily: 'monospace', fontWeight: 500 }}>v{version}</span>
      ),
    },
    {
      title: '状态',
      key: 'status',
      width: 100,
      render: (_: unknown, record: RemoteVersion) =>
        installedSet.has(record.version) ? (
          <Tag color="success">已安装</Tag>
        ) : null,
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_: unknown, record: RemoteVersion) => (
        <Button
          type="primary"
          size="small"
          icon={<DownloadOutlined />}
          disabled={installedSet.has(record.version) || !!installingVersion}
          loading={installingVersion === record.version}
          onClick={() => handleInstall(record.version)}
        >
          {installingVersion === record.version ? '安装中' : '安装'}
        </Button>
      ),
    },
  ];

  const tabItems = [
    {
      key: 'select',
      label: '选择版本',
      children: (
        <Spin spinning={loadingAvailable} tip="正在获取可用版本...">
          <div style={{ marginBottom: 12, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Text type="secondary" style={{ fontSize: 12 }}>
              显示最新的可用版本，点击安装按钮进行安装
            </Text>
            <Button
              size="small"
              icon={<ReloadOutlined />}
              onClick={fetchAvailableVersions}
              loading={loadingAvailable}
            >
              刷新
            </Button>
          </div>
          <Table
            columns={columns}
            dataSource={availableVersions}
            rowKey="version"
            size="small"
            pagination={{ pageSize: 8, size: 'small' }}
            scroll={{ y: 280 }}
          />
        </Spin>
      ),
    },
    {
      key: 'manual',
      label: '手动输入',
      children: (
        <Form form={form} layout="vertical">
          <Form.Item
            name="version"
            label="版本号"
            rules={[
              { required: true, message: '请输入版本号' },
              {
                pattern: versionRegex,
                message: '版本号格式错误，请输入 x.y.z 格式（如 20.19.0）',
              },
            ]}
          >
            <Input
              placeholder="例如：20.19.0、18.20.8、16.20.2"
              disabled={!!installingVersion}
              autoFocus
            />
          </Form.Item>

          <div style={{ marginTop: 8 }}>
            <Text type="secondary" style={{ fontSize: 12 }}>
              提示：如果列表中没有你需要的版本，可以在这里手动输入版本号
            </Text>
          </div>
        </Form>
      ),
    },
  ];

  return (
    <Modal
      title="安装 Node.js 版本"
      open={open}
      onOk={handleOk}
      onCancel={handleCancel}
      okText={activeTab === 'manual' ? '安装' : undefined}
      cancelText="关闭"
      confirmLoading={!!installingVersion}
      closable={!installingVersion}
      maskClosable={!installingVersion}
      width={500}
      footer={activeTab === 'select' ? (
        <Button onClick={handleCancel}>关闭</Button>
      ) : undefined}
    >
      {error && (
        <Alert
          message={error}
          type="error"
          showIcon
          style={{ marginBottom: 16 }}
          closable
          onClose={() => setError(null)}
        />
      )}

      {installingVersion && (
        <Alert
          message={`正在安装 Node.js ${installingVersion}...`}
          description="安装过程可能需要几分钟，请耐心等待。"
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}

      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={tabItems}
        size="small"
      />
    </Modal>
  );
};

export default InstallModal;
