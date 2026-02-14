package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// DownloadStockVideo 搜索并下载视频
func DownloadStockVideo(keyword string, filename string) error {
	PexelsApiKey := os.Getenv("PEXELS_API_KEY")
	if PexelsApiKey == "" {
		return fmt.Errorf("未设置 PEXELS_API_KEY 环境变量")
	}
	// 1. URL 编码 (解决空格导致崩溃的问题)
	// "compound interest" -> "compound+interest" 或 "compound%20interest"
	safeKeyword := url.QueryEscape(keyword)

	// 2. 构造 URL
	apiURL := fmt.Sprintf("https://api.pexels.com/videos/search?query=%s&orientation=portrait&per_page=1&size=medium", safeKeyword)

	// 3. 创建请求 (增加了错误检查)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("构造请求失败(可能是关键词非法): %v", err)
	}

	req.Header.Set("Authorization", PexelsApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Pexels API 报错: %d (请检查 API Key 是否正确)", resp.StatusCode)
	}

	var result struct {
		Videos []struct {
			VideoFiles []struct {
				Link    string `json:"link"`
				Quality string `json:"quality"`
				Width   int    `json:"width"`
			} `json:"video_files"`
		} `json:"videos"`
	}

	body, _ := io.ReadAll(resp.Body)
	// 忽略 JSON 解析错误，因为有时候 body 为空
	_ = json.Unmarshal(body, &result)

	if len(result.Videos) == 0 {
		// 有时候搜不到，我们可以尝试把空格拆开搜第一个词，或者直接跳过
		return fmt.Errorf("没找到关于 [%s] 的视频", keyword)
	}

	// 优先找高清(hd)且宽度适中的链接
	downloadUrl := ""
	for _, f := range result.Videos[0].VideoFiles {
		// Pexels 有时候会给 4k 视频，太大了，限制一下宽度
		if f.Quality == "hd" && f.Width < 2500 {
			downloadUrl = f.Link
			break
		}
	}
	// 兜底
	if downloadUrl == "" && len(result.Videos[0].VideoFiles) > 0 {
		downloadUrl = result.Videos[0].VideoFiles[0].Link
	}

	if downloadUrl == "" {
		return fmt.Errorf("虽然找到了视频对象，但没有可下载的链接")
	}

	fmt.Printf("⬇️ 正在下载素材 [%s]...\n", keyword)
	return downloadFile(downloadUrl, filename)
}

func downloadFile(url string, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
