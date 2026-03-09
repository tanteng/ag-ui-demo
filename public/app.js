// AG-UI 协议演示 - SSE 流式版本

// AG UI: 执行任务 - 流式
async function executeTask() {
  const goal = document.getElementById('goalInput').value.trim();
  if (!goal) {
    alert('请输入目标');
    return;
  }
  
  const resultDiv = document.getElementById('aguiResult');
  resultDiv.style.display = 'block';
  
  let html = `
    <div class="goal-display">
      <div class="label">🎯 目标</div>
      <div class="goal">${goal}</div>
    </div>
    <div class="steps" id="stepsContainer">
      <p style="color:#ffc800;">⏳ 等待 AI 响应...</p>
    </div>
  `;
  resultDiv.innerHTML = html;
  
  const stepsContainer = document.getElementById('stepsContainer');
  
  // 调试区域
  const debugDiv = document.getElementById('debugEvents');
  if (debugDiv) {
    debugDiv.style.display = 'block';
    debugDiv.innerHTML = '';
  }
  
  try {
    const response = await fetch('/api/task/stream', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ goal })
    });
    
    if (!response.ok) {
      throw new Error('请求失败: ' + response.status);
    }
    
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';
    
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      
      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop();
      
      for (const line of lines) {
        if (!line.trim() || !line.startsWith('data:')) continue;
        
        const dataStr = line.substring(5).trim();
        if (!dataStr) continue;
        
        try {
          const data = JSON.parse(dataStr);
          const eventType = data.type || 'unknown';
          
          // 显示调试
          if (debugDiv) {
            const eventDiv = document.createElement('div');
            eventDiv.style.cssText = 'padding:4px 8px;margin:2px 0;background:rgba(0,0,0,0.3);border-radius:4px;font-size:11px;';
            eventDiv.textContent = eventType;
            debugDiv.appendChild(eventDiv);
          }
          
          handleEvent(eventType, data, stepsContainer);
        } catch (e) {
          console.error('解析错误:', e);
        }
      }
    }
  } catch (error) {
    stepsContainer.innerHTML += `<p style="color:#ff6b6b;">❌ 错误: ${error.message}</p>`;
  }
}

function handleEvent(type, data, container) {
  console.log('处理事件:', type, data);
  
  // 过滤乱码字符
  const cleanText = (text) => text.replace(/\ufffd/g, '').trim();
  
  switch (type) {
    case 'RUN_STARTED':
      container.innerHTML = `<p style="color:#00d4ff;">🚀 任务开始</p>`;
      break;
      
    case 'REASONING_START':
      container.innerHTML += `<p style="color:#9b59b6;">🧠 开始推理...</p>`;
      break;
      
    case 'REASONING_MESSAGE_CONTENT':
      container.innerHTML += `<p style="color:#9b59b6;">💭 ${cleanText(data.delta)}</p>`;
      break;
      
    case 'REASONING_END':
      break;
      
    case 'STEP_STARTED':
      container.innerHTML += `
        <div class="step active">
          <div class="step-number">▶</div>
          <div class="step-text">${data.stepName}</div>
          <div class="step-status">进行中</div>
        </div>
      `;
      break;
      
    case 'STEP_FINISHED':
      // 更新步骤显示
      const steps = container.querySelectorAll('.step');
      for (const step of steps) {
        if (step.querySelector('.step-text').textContent === data.stepName) {
          step.classList.remove('active');
          step.classList.add('completed');
          step.querySelector('.step-number').textContent = '✓';
          step.querySelector('.step-status').textContent = '完成';
        }
      }
      break;
      
    case 'TEXT_MESSAGE_CONTENT':
      // 显示最终结果
      let existingResult = container.querySelector('.result');
      if (!existingResult) {
        existingResult = document.createElement('div');
        existingResult.className = 'result';
        existingResult.style.cssText = 'background:rgba(0,255,136,0.1);border-left:4px solid #00ff88;margin-top:16px;padding:16px;border-radius:10px;';
        container.appendChild(existingResult);
      }
      existingResult.innerHTML += `<span style="color:#ccc;white-space:pre-wrap;">${cleanText(data.delta)}</span>`;
      break;
      
    case 'TEXT_MESSAGE_END':
      const resultDiv = container.querySelector('.result');
      if (resultDiv) {
        resultDiv.innerHTML = `<h3 style="color:#00ff88;margin-bottom:12px;">✅ 完成</h3>` + resultDiv.innerHTML;
      }
      break;
      
    case 'RUN_FINISHED':
      container.innerHTML += `
        <div class="note" style="margin-top:20px;">
          💡 <strong>AG-UI 协议事件流已结束</strong><br>
          执行时间: ${data.result?.executionTime || 'N/A'}<br>
          步骤数: ${data.result?.steps || 0}
        </div>
      `;
      break;
      
    case 'RUN_ERROR':
      container.innerHTML += `<p style="color:#ff6b6b;">❌ 错误: ${data.message}</p>`;
      break;
  }
  
  container.scrollTop = container.scrollHeight;
}

// A2UI: 决策分析
async function showDecision() {
  const goal = document.getElementById('decisionInput').value.trim();
  if (!goal) {
    alert('请输入决策问题');
    return;
  }
  
  const resultDiv = document.getElementById('a2uiResult');
  resultDiv.style.display = 'block';
  resultDiv.innerHTML = `
    <div class="goal-display">
      <div class="label">❓ 问题</div>
      <div class="goal">${goal}</div>
    </div>
    <div id="decisionContainer">
      <p style="color:#ffc800;">⏳ 等待 AI 决策分析...</p>
    </div>
  `;
  
  const container = document.getElementById('decisionContainer');
  
  try {
    const response = await fetch('/api/decision/stream?goal=' + encodeURIComponent(goal));
    if (!response.ok) {
      throw new Error('请求失败: ' + response.status);
    }
    
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';
    
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      
      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop();
      
      for (const line of lines) {
        if (!line.trim() || !line.startsWith('data:')) continue;
        
        const dataStr = line.substring(5).trim();
        if (!dataStr) continue;
        
        try {
          const data = JSON.parse(dataStr);
          handleDecisionEvent(data.type, data, container);
        } catch (e) {
          console.error('解析错误:', e);
        }
      }
    }
  } catch (error) {
    container.innerHTML += `<p style="color:#ff6b6b;">❌ 错误: ${error.message}</p>`;
  }
}

function handleDecisionEvent(type, data, container) {
  if (type === 'RUN_STARTED') {
    container.innerHTML = `<p style="color:#00d4ff;">🚀 分析中...</p>`;
  } else if (type === 'TEXT_MESSAGE_CONTENT') {
    let existingResult = container.querySelector('.result');
    if (!existingResult) {
      existingResult = document.createElement('div');
      existingResult.className = 'result';
      existingResult.style.cssText = 'background:rgba(255,200,0,0.1);border-left:4px solid #ffc800;margin-top:16px;padding:16px;border-radius:10px;';
      container.appendChild(existingResult);
    }
    existingResult.innerHTML += `<span style="color:#ccc;white-space:pre-wrap;">${data.delta}</span>`;
  } else if (type === 'RUN_FINISHED') {
    container.innerHTML += `
      <div class="note" style="margin-top:20px;">
        💡 <strong>决策分析完成</strong>
      </div>
    `;
  }
}

// 回车触发
document.getElementById('goalInput').addEventListener('keypress', (e) => {
  if (e.key === 'Enter') executeTask();
});

document.getElementById('decisionInput').addEventListener('keypress', (e) => {
  if (e.key === 'Enter') showDecision();
});
