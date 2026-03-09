import React, { useState } from 'react';
import { Button, Card, Empty, message, Popconfirm, Space, Table, Tag } from 'antd';
import {
  CheckCircleFilled,
  DeleteOutlined,
  PlusOutlined,
  ReloadOutlined,
  SwapOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import InstallModal from './InstallModal';
import { NodeVersion, useNvmStore } from '../stores/nvmStore';

const VersionList: React.FC = () => {
  const [installModalOpen, setInstallModalOpen] = useState(false);
  const { versions, loading, isNvmAvailable, refreshAll, uninstallVersion, useVersion } = useNvmStore();

  const handleUse = async (version: string) => {
    const success = await useVersion(version);
    if (success) {
      message.success(`已切换到 Node.js ${version}`);
    }
  };

  const handleUninstall = async (version: string) => {
    const success = await uninstallVersion(version);
    if (success) {
      message.success(`Node.js ${version} 已卸载`);
    }
  };

  const columns: ColumnsType<NodeVersion> = [
    {
      title: '版本',
      dataIndex: 'version',
      key: 'version',
      render: (version: string, record: NodeVersion) => (
        <Space>
          <span style={{ fontFamily: 'monospace', fontWeight: 500 }}>v{version}</span>
          {record.isCurrent && (
            <Tag color="success" icon={<CheckCircleFilled />}>
              当前
            </Tag>
          )}
        </Space>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 220,
      render: (_value, record) => (
        <Space size="small">
          <Button
            type="primary"
            size="small"
            icon={<SwapOutlined />}
            disabled={record.isCurrent || loading}
            onClick={() => handleUse(record.version)}
          >
            切换
          </Button>
          <Popconfirm
            title="确认卸载"
            description={`确定要卸载 Node.js ${record.version} 吗？`}
            onConfirm={() => handleUninstall(record.version)}
            okText="确定"
            cancelText="取消"
            disabled={record.isCurrent}
          >
            <Button
              danger
              size="small"
              icon={<DeleteOutlined />}
              disabled={record.isCurrent || loading}
            >
              卸载
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      <Card
        title="已安装版本"
        size="small"
        extra={
          <Space>
            <Button
              icon={<ReloadOutlined />}
              onClick={refreshAll}
              loading={loading}
              disabled={!isNvmAvailable}
            >
              刷新
            </Button>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => setInstallModalOpen(true)}
              disabled={!isNvmAvailable}
            >
              安装新版本
            </Button>
          </Space>
        }
        style={{ marginBottom: 16 }}
      >
        <Table
          columns={columns}
          dataSource={versions}
          rowKey="version"
          loading={loading}
          pagination={false}
          size="small"
          locale={{
            emptyText: (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description={isNvmAvailable ? '暂无已安装的 Node.js 版本' : 'NVM 不可用'}
              />
            ),
          }}
        />
      </Card>

      <InstallModal open={installModalOpen} onClose={() => setInstallModalOpen(false)} />
    </>
  );
};

export default VersionList;
