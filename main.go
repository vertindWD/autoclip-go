package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

func main() {
	// ================= 配置区 =================
	deepSeekKey := os.Getenv("DEEPSEEK_API_KEY")
	if deepSeekKey == "" {
		log.Fatal("❌ 错误: 未检测到环境变量 DEEPSEEK_API_KEY")
	}
	topic := "复利的威力"
	outputDir := "output" // 固定输出文件夹
	// ==========================================

	// 0. 准备工作：确保输出目录存在
	_ = os.MkdirAll(outputDir, 0755)

	// 1. 策划阶段 (AI)
	fmt.Printf("🧠 正在请求 AI 策划 300 字视频: [%s]...\n", topic)
	plan, err := GeneratePlan(deepSeekKey, topic)
	if err != nil {
		log.Fatalf("❌ 策划失败: %v", err)
	}

	// 2. 配音阶段 (TTS)
	audioPath := filepath.Join(outputDir, "audio.mp3")
	fmt.Println("🎙️ 正在生成纯净语音...")
	if err := generateAudio(audioPath, plan.Script); err != nil {
		log.Fatalf("❌ 配音失败: %v", err)
	}

	// 3. 素材准备 (并发下载)
	fmt.Println("📥 正在并发下载视频素材...")
	limit := 8
	videoList := make([]string, limit)
	var wg sync.WaitGroup
	sem := make(chan struct{}, 3)

	for i := 0; i < limit; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			keyword := plan.Keywords[idx%len(plan.Keywords)]
			filename := filepath.Join(outputDir, fmt.Sprintf("temp_clip_%d.mp4", idx))
			if err := DownloadStockVideo(keyword, filename); err != nil {
				return
			}
			videoList[idx] = filename
		}(i)
	}
	wg.Wait()

	// 4. 合成阶段
	finalVideoPath := filepath.Join(outputDir, "final_movie.mp4")
	fmt.Println("🎬 正在合成: 文字结束即切断视频...")

	var validClips []string
	for _, v := range videoList {
		if v != "" {
			validClips = append(validClips, v)
		}
	}

	err = generateFinalVideo(validClips, audioPath, finalVideoPath)
	if err != nil {
		log.Fatalf("❌ 合成失败: %v", err)
	}

	// 5. 自动粉碎中间文件
	fmt.Println("🧹 正在清理没用的临时文件...")
	CleanTempFiles(validClips, audioPath)

	fmt.Printf("\n🎉 生成成功！请查看: %s\n", finalVideoPath)
}

func generateAudio(path, text string) error {
	return exec.Command("python3", "-m", "edge_tts",
		"--text", text, "--voice", "zh-CN-YunxiNeural", "--write-media", path,
	).Run()
}

func generateFinalVideo(videoInputs []string, audioPath, outPath string) error {
	args := []string{}
	for _, v := range videoInputs {
		args = append(args, "-i", v)
	}
	args = append(args, "-i", audioPath)
	audioIdx := len(videoInputs)

	filterComplex := ""
	concatIn := ""
	for i := 0; i < len(videoInputs); i++ {
		// 强制 1080x1920 竖屏，30帧，统一格式以防止报错
		filterComplex += fmt.Sprintf("[%d:v]scale=1080:1920:force_original_aspect_ratio=increase,crop=1080:1920,setsar=1,fps=30,format=yuv420p[v%d];", i, i)
		concatIn += fmt.Sprintf("[v%d]", i)
	}
	filterComplex += fmt.Sprintf("%sconcat=n=%d:v=1:a=0[vcat]", concatIn, len(videoInputs))

	args = append(args,
		"-filter_complex", filterComplex,
		"-map", "[vcat]", "-map", fmt.Sprintf("%d:a", audioIdx),
		"-c:v", "libx264", "-preset", "fast", "-c:a", "aac",
		"-shortest", // ⭐ 核心：配音音频结束，视频画面立即停止
		"-y", outPath,
	)
	return exec.Command("ffmpeg", args...).Run()
}
