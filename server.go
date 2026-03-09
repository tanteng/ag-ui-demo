package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// 从环境变量读取配置
func getConfig() (apiURL, apiKey string) {
	apiURL = os.Getenv("AG_UI_API_URL")
	if apiURL == "" {
		apiURL = "https://api.siliconflow.cn/v1/chat/completions"
	}
	
	apiKey = os.Getenv("AG_UI_API_KEY")
	return
}

var (
	logger *log.Logger
	logFile *os.File
)

// ========== AG-UI 事件类型常量 ==========

const (
	// 生命周期事件
	EVENT_RUN_STARTED   = "RUN_STARTED"
	EVENT_RUN_FINISHED  = "RUN_FINISHED"
	EVENT_RUN_ERROR     = "RUN_ERROR"
	EVENT_STEP_STARTED  = "STEP_STARTED"
	EVENT_STEP_FINISHED = "STEP_FINISHED"

	// 文本消息事件
	EVENT_TEXT_MESSAGE_START   = "TEXT_MESSAGE_START"
	EVENT_TEXT_MESSAGE_CONTENT = "TEXT_MESSAGE_CONTENT"
	EVENT_TEXT_MESSAGE_END     = "TEXT_MESSAGE_END"

	// 工具调用事件
	EVENT_TOOL_CALL_START  = "TOOL_CALL_START"
	EVENT_TOOL_CALL_ARGS   = "TOOL_CALL_ARGS"
	EVENT_TOOL_CALL_END   = "TOOL_CALL_END"
	EVENT_TOOL_CALL_RESULT = "TOOL_CALL_RESULT"

	// 推理事件
	EVENT_REASONING_START            = "REASONING_START"
	EVENT_REASONING_MESSAGE_START    = "REASONING_MESSAGE_START"
	EVENT_REASONING_MESSAGE_CONTENT  = "REASONING_MESSAGE_CONTENT"
	EVENT_REASONING_MESSAGE_END      = "REASONING_MESSAGE_END"
	EVENT_REASONING_END              = "REASONING_END"

	// 状态事件
	EVENT_STATE_SNAPSHOT = "STATE_SNAPSHOT"
	EVENT_STATE_DELTA    = "STATE_DELTA"
)

// ========== 日志初始化 ==========

func initLog() {
	os.MkdirAll("/var/log/ag-ui", 0755)
	today := time.Now().Format("2006-01-02")
	logPath := fmt.Sprintf("/var/log/ag-ui/server-%s.log", today)

	var err error
	logFile, err = os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("无法创建日志文件: %v", err)
		logger = log.New(os.Stdout, "[AG-UI] ", log.LstdFlags)
		return
	}

	logger = log.New(logFile, "[AG-UI] ", log.LstdFlags|log.Lshortfile)
	logger.Println("========== 服务启动 (AG-UI 协议) ==========")
}

// ========== SSE 发送事件 (符合 AG-UI 规范) ==========

func sendSSEvent(w http.ResponseWriter, eventType string, data map[string]interface{}) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// AG-UI 标准格式: 包含 type 和 timestamp
	event := map[string]interface{}{
		"type":      eventType,
		"timestamp": time.Now().UnixMilli(),
	}

	// 合并数据
	for k, v := range data {
		event[k] = v
	}

	jsonData, _ := json.Marshal(event)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

// ========== AI 调用 ==========

func callAI(prompt string) (string, error) {
	apiURL, apiKey := getConfig()
	
	if apiKey == "" {
		return "", fmt.Errorf("请设置 AG_UI_API_KEY 环境变量")
	}
	
	reqBody := map[string]interface{}{
		"model": "Qwen/Qwen2.5-7B-Instruct",
		"messages": []map[string]string{
			{"role": "system", "content": "你是一个专业的AI助手。请用JSON格式回复。"},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error: %s", string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	choices := result["choices"].([]interface{})
	firstChoice := choices[0].(map[string]interface{})
	message := firstChoice["message"].(map[string]interface{})
	content := message["content"].(string)

	return content, nil
}

// ========== 流式任务处理 (完全符合 AG-UI 规范) ==========

func handleTaskStream(w http.ResponseWriter, goal string) {
	runID := fmt.Sprintf("run-%d", time.Now().Unix())
	threadID := "default"
	messageID := fmt.Sprintf("msg-%d", time.Now().Unix())
	startTime := time.Now()

	// 1. RUN_STARTED - 任务开始 (必需)
	sendSSEvent(w, EVENT_RUN_STARTED, map[string]interface{}{
		"threadId": threadID,
		"runId":    runID,
		"input":    map[string]string{"goal": goal},
	})
	logger.Printf("任务开始: %s - %s", runID, goal)

	// 2. REASONING_START - 开始推理
	sendSSEvent(w, EVENT_REASONING_START, map[string]interface{}{
		"messageId": messageID,
	})

	// 3. TEXT_MESSAGE_START - 消息开始
	sendSSEvent(w, EVENT_TEXT_MESSAGE_START, map[string]interface{}{
		"messageId": messageID,
		"role":      "assistant",
	})

	// 4. 分析任务 - 推理阶段
	sendSSEvent(w, EVENT_REASONING_MESSAGE_START, map[string]interface{}{
		"messageId": messageID,
		"role":      "assistant",
	})

	// 发送思考内容
	thinkingContent := "正在分析任务需求..."
	sendSSEvent(w, EVENT_REASONING_MESSAGE_CONTENT, map[string]interface{}{
		"messageId": messageID,
		"delta":     thinkingContent,
	})

	prompt := `用户目标: ` + goal + `
请分析这个任务，给出：
1. 任务类型（code/data_analysis/information_search/content_summary/translation/plan/other）
2. 执行步骤（3-5个步骤，用|分隔）
3. 任务详情（一句话描述）

请用以下JSON格式回复：
{"type":"任务类型","steps":["步骤1","步骤2","步骤3"],"details":"任务详情"}`

	result, err := callAI(prompt)
	if err != nil {
		sendSSEvent(w, EVENT_RUN_ERROR, map[string]interface{}{
			"message": "任务分析失败: " + err.Error(),
			"code":    "ANALYSIS_ERROR",
		})
		return
	}

	// 思考结束
	sendSSEvent(w, EVENT_REASONING_MESSAGE_END, map[string]interface{}{
		"messageId": messageID,
	})
	sendSSEvent(w, EVENT_REASONING_END, map[string]interface{}{
		"messageId": messageID,
	})

	// 解析任务分析
	var taskInfo struct {
		Type   string   `json:"type"`
		Steps  []string `json:"steps"`
		Details string  `json:"details"`
	}
	json.Unmarshal([]byte(result), &taskInfo)

	// 5. 步骤执行
	for idx := range taskInfo.Steps {
		sendSSEvent(w, EVENT_STEP_STARTED, map[string]interface{}{
			"stepName": fmt.Sprintf("step-%d", idx+1),
		})
		time.Sleep(200 * time.Millisecond)
		sendSSEvent(w, EVENT_STEP_FINISHED, map[string]interface{}{
			"stepName": fmt.Sprintf("step-%d", idx+1),
		})
	}

	// 6. 生成结果
	sendSSEvent(w, EVENT_STEP_STARTED, map[string]interface{}{
		"stepName": "generate",
	})

	resultPrompt := fmt.Sprintf(`用户需求: %s
任务类型: %s
请完成这个任务，给出完整的结果。`, goal, taskInfo.Type)

	finalResult, err := callAI(resultPrompt)
	if err != nil {
		sendSSEvent(w, EVENT_RUN_ERROR, map[string]interface{}{
			"message": "结果生成失败: " + err.Error(),
			"code":    "GENERATION_ERROR",
		})
		return
	}

	sendSSEvent(w, EVENT_STEP_FINISHED, map[string]interface{}{
		"stepName": "generate",
	})

	// 7. 流式发送消息内容 (TEXT_MESSAGE_CONTENT)
	content := strings.Trim(finalResult, "` \n")
	chunkSize := 40
	for i := 0; i < len(content); i += chunkSize {
		end := i + chunkSize
		if end > len(content) {
			end = len(content)
		}
		sendSSEvent(w, EVENT_TEXT_MESSAGE_CONTENT, map[string]interface{}{
			"messageId": messageID,
			"delta":     content[i:end],
		})
		time.Sleep(50 * time.Millisecond)
	}

	// 8. TEXT_MESSAGE_END - 消息结束
	sendSSEvent(w, EVENT_TEXT_MESSAGE_END, map[string]interface{}{
		"messageId": messageID,
	})

	// 9. RUN_FINISHED - 任务完成 (必需)
	elapsed := time.Since(startTime)
	sendSSEvent(w, EVENT_RUN_FINISHED, map[string]interface{}{
		"threadId": threadID,
		"runId":    runID,
		"result": map[string]interface{}{
			"executionTime": elapsed.String(),
			"steps":        len(taskInfo.Steps),
			"taskType":     taskInfo.Type,
		},
	})

	logger.Printf("任务完成: %s - 耗时 %v", runID, elapsed)
}

// ========== 决策流式处理 ==========

func handleDecisionStream(w http.ResponseWriter, goal string) {
	runID := fmt.Sprintf("decision-%d", time.Now().Unix())
	threadID := "decision-thread"
	messageID := fmt.Sprintf("msg-decision-%d", time.Now().Unix())

	// RUN_STARTED
	sendSSEvent(w, EVENT_RUN_STARTED, map[string]interface{}{
		"threadId": threadID,
		"runId":    runID,
		"input":    map[string]string{"goal": goal},
	})

	// TEXT_MESSAGE_START
	sendSSEvent(w, EVENT_TEXT_MESSAGE_START, map[string]interface{}{
		"messageId": messageID,
		"role":      "assistant",
	})

	// REASONING
	sendSSEvent(w, EVENT_REASONING_START, map[string]interface{}{
		"messageId": messageID,
	})
	sendSSEvent(w, EVENT_REASONING_MESSAGE_START, map[string]interface{}{
		"messageId": messageID,
	})
	sendSSEvent(w, EVENT_REASONING_MESSAGE_CONTENT, map[string]interface{}{
		"messageId": messageID,
		"delta":     "正在分析决策选项...",
	})
	sendSSEvent(w, EVENT_REASONING_MESSAGE_END, map[string]interface{}{
		"messageId": messageID,
	})
	sendSSEvent(w, EVENT_REASONING_END, map[string]interface{}{
		"messageId": messageID,
	})

	prompt := `用户问题: ` + goal + `
请分析并给出决策建议。

请用以下JSON格式回复：
{"decision":"最终决策","reasoning":["理由1","理由2","理由3"],"alternatives":[{"name":"方案A","score":80,"reason":"原因"},{"name":"方案B","score":70,"reason":"原因"}],"confidence":0.85,"uncertainties":["不确定因素1","不确定因素2"]}`

	result, err := callAI(prompt)
	if err != nil {
		sendSSEvent(w, EVENT_RUN_ERROR, map[string]interface{}{
			"message": "决策分析失败: " + err.Error(),
		})
		return
	}

	result = strings.Trim(result, "` \n")
	if strings.HasPrefix(result, "json") {
		result = strings.TrimPrefix(result, "json")
	}

	// 流式发送
	for i := 0; i < len(result); i += 30 {
		end := i + 30
		if end > len(result) {
			end = len(result)
		}
		sendSSEvent(w, EVENT_TEXT_MESSAGE_CONTENT, map[string]interface{}{
			"messageId": messageID,
			"delta":     result[i:end],
		})
		time.Sleep(30 * time.Millisecond)
	}

	// TEXT_MESSAGE_END
	sendSSEvent(w, EVENT_TEXT_MESSAGE_END, map[string]interface{}{
		"messageId": messageID,
	})

	// RUN_FINISHED
	sendSSEvent(w, EVENT_RUN_FINISHED, map[string]interface{}{
		"threadId": threadID,
		"runId":    runID,
		"result":   map[string]string{"status": "completed"},
	})
}

// ========== 主函数 ==========

func main() {
	initLog()

	// SSE 流式端点 - 任务执行
	http.HandleFunc("/api/task/stream", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Goal string `json:"goal"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		handleTaskStream(w, req.Goal)
	})

	// SSE 流式端点 - 决策分析
	http.HandleFunc("/api/decision/stream", func(w http.ResponseWriter, r *http.Request) {
		goal := r.URL.Query().Get("goal")
		if goal == "" {
			goal = "一般决策"
		}
		handleDecisionStream(w, goal)
	})

	// 静态文件
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	logger.Println("AG-UI 协议演示服务器运行在 http://localhost:3000")
	logger.Println("遵循 AG-UI 官方协议规范 (EventType)")
	logger.Fatal(http.ListenAndServe(":3000", nil))
}
