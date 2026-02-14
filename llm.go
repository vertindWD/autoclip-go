package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

type VideoPlan struct {
	Script   string   `json:"script"`
	Keywords []string `json:"keywords"`
}

func GeneratePlan(apiKey string, topic string) (*VideoPlan, error) {
	systemPrompt := `你是一个短视频导演。
1. script: 必须生成 300 字左右的口播文案，节奏快。
2. keywords: 生成 8 个英文关键词。
必须直接返回纯 JSON，禁止任何 Markdown 格式。`

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
	json.Unmarshal(body, &aiResp)

	raw := aiResp.Choices[0].Message.Content
	raw = strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(raw), "```"), "```json")

	var plan VideoPlan
	json.Unmarshal([]byte(raw), &plan)
	return &plan, nil
}
