import React from 'react';
import { Card, List, Tag, Button, Empty } from 'antd';
import { ClearOutlined } from '@ant-design/icons';
import { useNvmStore, LogEntry } from '../stores/nvmStore';

const levelColorMap: Record<LogEntry['level'], string> = {
  info: 'processing',
  success: 'success',
  warning: 'warning',
  error: 'error',
};

const levelTextMap: Record<LogEntry['level'], string> = {
  info: '信息',
  success: '成功',
  warning: '警告',
  error: '错误',
};

const LogPanel: React.FC = () => {
  const { logs, clearLogs } = useNvmStore();

  return (
    <Card
      title="操作日志"
      size="small"
      extra={
        <Button
          size="small"
          icon={<ClearOutlined />}
          onClick={clearLogs}
          disabled={logs.length === 0}
        >
          清空
        </Button>
      }
      bodyStyle={{
        maxHeight: 200,
        overflowY: 'auto',
        padding: logs.length === 0 ? 24 : '8px 16px',
      }}
    >
      {logs.length === 0 ? (
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description="暂无日志"
          style={{ margin: 0 }}
        />
      ) : (
        <div>
          {logs.map((log) => (
            <div
              key={log.id}
              style={{
                display: 'flex',
                alignItems: 'center',
                padding: '6px 0',
                borderBottom: '1px solid #f0f0f0',
              }}
            >
              <span
                style={{
                  fontFamily: 'monospace',
                  fontSize: 12,
                  color: '#999',
                  width: 80,
                  flexShrink: 0,
                }}
              >
                [{log.time}]
              </span>
              <Tag
                color={levelColorMap[log.level]}
                style={{ width: 48, textAlign: 'center', flexShrink: 0 }}
              >
                {levelTextMap[log.level]}
              </Tag>
              <span style={{ fontSize: 13, marginLeft: 8, flex: 1 }}>
                {log.message}
              </span>
            </div>
          ))}
        </div>
      )}
    </Card>
  );
};

export default LogPanel;
