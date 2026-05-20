import React from 'react'
import {createRoot} from 'react-dom/client'
import { Alert, Button, Space } from 'antd'
import './style.css'
import App from './App'

interface ErrorBoundaryProps {
    children: React.ReactNode
}

interface ErrorBoundaryState {
    hasError: boolean
    errorMessage: string
}

class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
    constructor(props: ErrorBoundaryProps) {
        super(props)
        this.state = {
            hasError: false,
            errorMessage: '',
        }
    }

    static getDerivedStateFromError(error: Error): ErrorBoundaryState {
        return {
            hasError: true,
            errorMessage: error.message || '未知错误',
        }
    }

    componentDidCatch(error: Error): void {
        console.error('App render failed:', error)
    }

    private handleRetry = () => {
        this.setState({
            hasError: false,
            errorMessage: '',
        })
    }

    render() {
        if (this.state.hasError) {
            return (
                <div style={{ padding: 24 }}>
                    <Alert
                        message="界面加载失败"
                        description={this.state.errorMessage}
                        type="error"
                        showIcon
                        action={
                            <Space>
                                <Button size="small" onClick={this.handleRetry}>
                                    重试
                                </Button>
                            </Space>
                        }
                    />
                </div>
            )
        }

        return this.props.children
    }
}

const container = document.getElementById('root')

const root = createRoot(container!)

root.render(
    <ErrorBoundary>
        <App/>
    </ErrorBoundary>
)
