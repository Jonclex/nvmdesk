import React from 'react';
import { Button, Card, Empty, message, Popconfirm, Space, Table, Tag, Tooltip, Typography } from 'antd';
import { AppstoreOutlined, DeleteOutlined, ReloadOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { GlobalNpmPackage, useNvmStore } from '../stores/nvmStore';

const { Text } = Typography;

const NpmPackageList: React.FC = () => {
  const {
    currentInfo,
    deletingPackage,
    fetchGlobalPackages,
    globalPackages,
    isNvmAvailable,
    loadingPackages,
    uninstallGlobalPackage,
  } = useNvmStore();

  const handleUninstall = async (name: string) => {
    const success = await uninstallGlobalPackage(name);
    if (success) {
      message.success(`已卸载全局 npm 包 ${name}`);
    }
  };

  const columns: ColumnsType<GlobalNpmPackage> = [
    {
      title: '包名',
      dataIndex: 'name',
      key: 'name',
      render: (name: string, record) => (
        <Space direction="vertical" size={0}>
          <Text strong style={{ fontFamily: 'monospace' }}>
            {name}
          </Text>
          <Tooltip title={record.path}>
            <Text type="secondary" ellipsis style={{ maxWidth: 420 }}>
              {record.path}
            </Text>
          </Tooltip>
        </Space>
      ),
    },
    {
      title: '版本',
      dataIndex: 'version',
      key: 'version',
      width: 140,
      render: (version: string) => <Tag color="blue">{version || '-'}</Tag>,
    },
    {
      title: '体积',
      dataIndex: 'sizeLabel',
      key: 'sizeLabel',
      width: 120,
      sorter: (a, b) => a.sizeBytes - b.sizeBytes,
      defaultSortOrder: 'descend',
      render: (sizeLabel: string) => <Text strong>{sizeLabel}</Text>,
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_value, record) => (
        <Popconfirm
          title="确认卸载"
          description={`确定要卸载全局 npm 包 ${record.name} 吗？`}
          okText="确定"
          cancelText="取消"
          onConfirm={() => handleUninstall(record.name)}
        >
          <Button
            danger
            size="small"
            icon={<DeleteOutlined />}
            loading={deletingPackage === record.name}
            disabled={loadingPackages || deletingPackage !== null}
          >
            删除
          </Button>
        </Popconfirm>
      ),
    },
  ];

  return (
    <Card
      title={
        <Space>
          <AppstoreOutlined />
          <span>当前 Node 全局 npm 包</span>
        </Space>
      }
      size="small"
      extra={
        <Space>
          <Text type="secondary">
            {currentInfo?.nodeVersion ? `Node.js v${currentInfo.nodeVersion}` : '未激活 Node.js'}
          </Text>
          <Button
            icon={<ReloadOutlined />}
            loading={loadingPackages}
            onClick={fetchGlobalPackages}
            disabled={!isNvmAvailable || deletingPackage !== null}
          >
            刷新
          </Button>
        </Space>
      }
      style={{ marginBottom: 16 }}
    >
      <Table
        columns={columns}
        dataSource={globalPackages}
        rowKey="name"
        loading={loadingPackages}
        pagination={{ pageSize: 8, showSizeChanger: false }}
        size="small"
        locale={{
          emptyText: (
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description={
                currentInfo?.nodeVersion
                  ? '当前 Node 环境下未发现全局 npm 包'
                  : '请先切换到可用的 Node.js 版本'
              }
            />
          ),
        }}
      />
    </Card>
  );
};

export default NpmPackageList;
