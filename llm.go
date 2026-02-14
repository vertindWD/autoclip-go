package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type VideoPlan struct {
	Script   string   `json:"script"`
	Keywords []string `json:"keywords"`
}

// GeneratePlan 现在同时接收字数和素材数要求
func GeneratePlan(apiKey string, topic string, wordCount, clipCount int) (*VideoPlan, error) {
	// 在 Prompt 中明确要求生成对应数量的关键词
	systemPrompt := fmt.Sprintf(`你是一个短视频导演。请根据主题执行：
1. script: 必须生成 %d 字左右的口播文案，节奏极快，适合 40 秒内念完。
2. keywords: 必须生成 %d 个不同的英文搜索关键词，每个词对应一个画面。
必须返回 JSON，不要包含任何 Markdown 标签或废话。`, wordCount, clipCount)

	reqBody := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": topic},
		},
	}
	jsonData, _ := json.Marshal(reqBody)

	client := &http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var aiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &aiResp); err != nil {
		return nil, fmt.Errorf("请求 API 失败: %v", err)
	}

	raw := aiResp.Choices[0].Message.Content

	// ⭐ 核心优化：鲁棒的 JSON 提取逻辑
	// 找到第一个 { 和最后一个 }，忽略掉 AI 的所有废话
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start == -1 || end == -1 || end < start {
		return nil, fmt.Errorf("AI 返回的不是合法的 JSON 格式: %s", raw)
	}
	cleanJson := raw[start : end+1]

	var plan VideoPlan
	if err := json.Unmarshal([]byte(cleanJson), &plan); err != nil {
		return nil, fmt.Errorf("解析策划方案失败: %v | 原始内容: %s", err, cleanJson)
	}
	return &plan, nil
}
