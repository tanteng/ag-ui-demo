# AG-UI Protocol Demo

基于 AG-UI 官方协议的流式事件演示项目。

## 功能特性

- ✅ 符合 AG-UI 官方协议规范
- ✅ SSE 流式事件推送
- ✅ 完整的事件类型支持
  - RUN_STARTED / RUN_FINISHED / RUN_ERROR
  - STEP_STARTED / STEP_FINISHED
  - TEXT_MESSAGE_START / TEXT_MESSAGE_CONTENT / TEXT_MESSAGE_END
  - REASONING_START / REASONING_MESSAGE_CONTENT / REASONING_END
  - TOOL_CALL_START / TOOL_CALL_ARGS / TOOL_CALL_END / TOOL_CALL_RESULT

## 快速开始

### 前置要求

- Go 1.18+
- API Key（硅基流动 / OpenAI / Claude）

### 1. 克隆项目

```bash
git clone git@github.com:tanteng/ag-ui-demo.git
cd ag-ui-demo
```

### 2. 配置环境变量

```bash
# 复制配置模板
cp .env.example .env

# 编辑 .env，填入你的 API Key
# 获取 API Key: https://cloud.siliconflow.cn
```

### 3. 运行

```bash
# Go 版本（推荐）
export AG_UI_API_KEY=your-api-key
go run server.go

# 或者 Node.js 版本（需要先安装依赖）
npm install
node server.js
```

### 4. 访问

打开浏览器: http://localhost:3000

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| AG_UI_API_URL | 硅基流动 API | AI API 端点 |
| AG_UI_API_KEY | - | AI API Key (必填) |

## 项目结构

```
ag-ui-demo/
├── server.go          # Go 后端 (符合 AG-UI 协议)
├── public/
│   ├── index.html    # 前端页面
│   └── app.js        # 前端逻辑
├── .env.example     # 环境变量模板
├── .gitignore
└── README.md
```

## API 接口

### 任务执行

```bash
curl -N -X POST http://localhost:3000/api/task/stream \
  -H "Content-Type: application/json" \
  -d '{"goal":"写一个登录功能"}'
```

### 决策分析

```bash
curl -N "http://localhost:3000/api/decision/stream?goal=选择哪个云服务器"
```

## 事件流示例

完整的 AG-UI 事件序列：

```
1. RUN_STARTED        → 任务开始
2. REASONING_START    → 开始推理
3. TEXT_MESSAGE_START → 消息开始
4. REASONING_MESSAGE_CONTENT → 推理内容（流式）
5. REASONING_MESSAGE_END → 推理结束
6. STEP_STARTED/STEP_FINISHED → 执行步骤
7. TEXT_MESSAGE_CONTENT → 最终结果（流式）
8. TEXT_MESSAGE_END → 消息结束
9. RUN_FINISHED → 任务完成
```

## 事件类型说明

| 事件类型 | 说明 |
|---------|------|
| RUN_STARTED | 任务开始 |
| RUN_FINISHED | 任务完成 |
| RUN_ERROR | 任务错误 |
| STEP_STARTED | 步骤开始 |
| STEP_FINISHED | 步骤完成 |
| TEXT_MESSAGE_START | 文本消息开始 |
| TEXT_MESSAGE_CONTENT | 文本消息内容（流式） |
| TEXT_MESSAGE_END | 文本消息结束 |
| REASONING_START | 推理开始 |
| REASONING_MESSAGE_CONTENT | 推理内容（流式） |
| REASONING_END | 推理结束 |
| TOOL_CALL_START | 工具调用开始 |
| TOOL_CALL_ARGS | 工具参数（流式） |
| TOOL_CALL_END | 工具调用结束 |
| TOOL_CALL_RESULT | 工具结果 |

## 部署

### Docker（可选）

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server .

FROM alpine
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 3000
CMD ["./server"]
```

### Systemd 服务

```ini
# /etc/systemd/system/ag-ui.service
[Unit]
Description=AG-UI Demo Server
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/ag-ui-demo
Environment=AG_UI_API_KEY=your-api-key
ExecStart=/home/ubuntu/ag-ui-demo/server
Restart=always

[Install]
WantedBy=multi-user.target
```

## 参考文档

- [AG-UI 官方文档](https://docs.ag-ui.com)
- [AG-UI GitHub](https://github.com/ag-ui-protocol/ag-ui)
- [AG-UI 协议规范](https://docs.ag-ui.com/concepts/events)
