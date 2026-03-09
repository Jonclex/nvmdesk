import React from 'react';
import { Button, Card, Empty, Space, Table, Tag, Tooltip, Typography } from 'antd';
import { AppstoreOutlined, ReloadOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { GlobalNpmPackage, useNvmStore } from '../stores/nvmStore';

const { Text } = Typography;

const NpmPackageList: React.FC = () => {
  const { currentInfo, fetchGlobalPackages, globalPackages, isNvmAvailable, loadingPackages } = useNvmStore();

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
            disabled={!isNvmAvailable}
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
