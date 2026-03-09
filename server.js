const express = require('express');
const path = require('path');

const app = express();
const PORT = 3000;

// 静态文件服务
app.use(express.static('public'));

// API 路由 - 模拟 AG UI 任务执行
app.post('/api/task', express.json(), (req, res) => {
  const { goal } = req.body;
  
  // 模拟 AI 任务执行
  const taskId = Date.now();
  
  // 返回任务结果 (AG UI 风格)
  res.json({
    taskId,
    status: 'completed',
    result: {
      message: `已完成目标: ${goal}`,
      actions: [
        '分析需求',
        '制定计划',
        '执行任务',
        '返回结果'
      ],
      executionTime: '2.3s'
    }
  });
});

// API 路由 - 模拟 A2UI 透明化
app.get('/api/decision/:id', (req, res) => {
  // 返回决策过程 (A2UI 风格)
  res.json({
    decisionId: req.params.id,
    decision: '选择方案 A',
    reasoning: [
      '方案A成本效益比最高',
      '实现难度适中',
      '用户满意度预测最高'
    ],
    alternatives: [
      { name: '方案B', score: 75, reason: '成本较高' },
      { name: '方案C', score: 60, reason: '实现周期长' }
    ],
    confidence: 0.85,
    uncertainties: [
      '市场环境可能变化',
      '技术实现存在风险'
    ],
    timestamp: new Date().toISOString()
  });
});

app.listen(PORT, () => {
  console.log(`AG UI & A2UI 演示服务器运行在 http://localhost:${PORT}`);
});
