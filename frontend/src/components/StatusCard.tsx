import React from 'react';
import { Card, Descriptions, Tag, Alert, Spin } from 'antd';
import { CheckCircleOutlined, CloseCircleOutlined, NodeIndexOutlined } from '@ant-design/icons';
import { useNvmStore } from '../stores/nvmStore';

const StatusCard: React.FC = () => {
  const { currentInfo, isNvmAvailable, loading } = useNvmStore();

  if (!isNvmAvailable) {
    return (
      <Alert
        message="NVM 未检测到"
        description="请先安装 nvm-windows 并确保其在系统 PATH 中，然后重启应用。"
        type="error"
        showIcon
        style={{ marginBottom: 16 }}
      />
    );
  }

  return (
    <Card
      title={
        <span>
          <NodeIndexOutlined style={{ marginRight: 8 }} />
          当前状态
        </span>
      }
      size="small"
      style={{ marginBottom: 16 }}
    >
      <Spin spinning={loading}>
        <Descriptions column={2} size="small">
          <Descriptions.Item label="Node.js">
            {currentInfo?.nodeVersion ? (
              <Tag color="green" icon={<CheckCircleOutlined />}>
                v{currentInfo.nodeVersion}
              </Tag>
            ) : (
              <Tag color="default" icon={<CloseCircleOutlined />}>
                未安装
              </Tag>
            )}
          </Descriptions.Item>
          <Descriptions.Item label="npm">
            {currentInfo?.npmVersion ? (
              <Tag color="blue">v{currentInfo.npmVersion}</Tag>
            ) : (
              <Tag color="default">-</Tag>
            )}
          </Descriptions.Item>
          <Descriptions.Item label="NVM">
            {currentInfo?.nvmVersion ? (
              <Tag color="purple">v{currentInfo.nvmVersion}</Tag>
            ) : (
              <Tag color="default">-</Tag>
            )}
          </Descriptions.Item>
          <Descriptions.Item label="NVM 目录">
            <span style={{ fontSize: 12, color: '#666', wordBreak: 'break-all' }}>
              {currentInfo?.nvmRoot || '-'}
            </span>
          </Descriptions.Item>
        </Descriptions>
      </Spin>
    </Card>
  );
};

export default StatusCard;
