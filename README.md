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

### 1. 克隆项目

```bash
git clone <your-repo-url>
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
# 方法一: 直接运行
export AG_UI_API_KEY=your-api-key
go run server.go

# 方法二: 使用 .env 文件
source .env
go run server.go
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
curl -X POST http://localhost:3000/api/task/stream \
  -H "Content-Type: application/json" \
  -d '{"goal":"写一个登录功能"}'
```

### 决策分析

```bash
curl "http://localhost:3000/api/decision/stream?goal=选择哪个云服务器"
```

## 参考文档

- [AG-UI 官方文档](https://docs.ag-ui.com)
- [AG-UI GitHub](https://github.com/ag-ui-protocol/ag-ui)
