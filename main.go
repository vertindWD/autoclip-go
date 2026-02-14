package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

func main() {
	// 1. 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ 未找到 .env 文件，将尝试从系统环境变量读取")
	}

	// 2. 读取配置并进行类型转换
	// 从环境变量读取字数要求，默认为 300
	wordCountStr := os.Getenv("SCRIPT_WORD_COUNT")
	wordCount, _ := strconv.Atoi(wordCountStr)
	if wordCount <= 0 {
		wordCount = 300
	}

	// 从环境变量读取素材下载数量，默认为 8
	clipCountStr := os.Getenv("VIDEO_CLIP_COUNT")
	clipCount, _ := strconv.Atoi(clipCountStr)
	if clipCount <= 0 {
		clipCount = 8
	}

	// 读取 API Key
	deepSeekKey := os.Getenv("DEEPSEEK_API_KEY")
	if deepSeekKey == "" {
		log.Fatal("❌ 错误: 请在 .env 或环境变量中设置 DEEPSEEK_API_KEY")
	}

	// 3. 解析命令行参数（视频主题）
	topicPtr := flag.String("t", "沉没成本的陷阱", "视频的主题内容")
	flag.Parse()
	topic := *topicPtr

	outputDir := "output"
	_ = os.MkdirAll(outputDir, 0755)

	// --- 执行流程 ---

	// 1. AI 策划 (动态传入字数)
	fmt.Printf("🧠 1. 正在策划视频: [%s] (要求: %d字 / %d个素材)...\n", topic, wordCount, clipCount)
	plan, err := GeneratePlan(deepSeekKey, topic, wordCount, clipCount)
	if err != nil {
		log.Fatalf("❌ 策划失败: %v", err)
	}

	// 2. 生成配音
	audioPath := filepath.Join(outputDir, "audio.mp3")
	fmt.Println("🎙️ 2. 正在生成配音...")
	if err := generateAudio(audioPath, plan.Script); err != nil {
		log.Fatalf("❌ 配音失败: %v", err)
	}

	// 3. 并发下载素材
	fmt.Printf("📥 3. 正在并行下载素材...\n")
	videoList := make([]string, clipCount)
	var wg sync.WaitGroup
	sem := make(chan struct{}, 3)

	for i := 0; i < clipCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// 如果 AI 返回的关键词够多，就按索引拿；不够则循环拿
			kwIdx := idx
			if kwIdx >= len(plan.Keywords) {
				kwIdx = idx % len(plan.Keywords)
			}
			keyword := plan.Keywords[kwIdx]

			filename := filepath.Join(outputDir, fmt.Sprintf("temp_%d.mp4", idx))
			if err := DownloadStockVideo(keyword, filename); err != nil {
				fmt.Printf("   ⚠️ 素材 [%s] 失败\n", keyword)
				return
			}
			videoList[idx] = filename
		}(i)
	}
	wg.Wait()

	// 4. 合成最终视频
	finalPath := filepath.Join(outputDir, "final_movie.mp4")
	fmt.Println("🎬 4. 正在合成 (强制以音频长度为准截断)...")

	var validClips []string
	for _, v := range videoList {
		if v != "" {
			validClips = append(validClips, v)
		}
	}

	if err := generateFinalVideo(validClips, audioPath, finalPath); err != nil {
		log.Fatalf("❌ 合成失败: %v", err)
	}

	// 5. 自动粉碎中间文件
	fmt.Println("Sweep 🧹 正在清理临时素材...")
	CleanTempFiles(validClips, audioPath)

	fmt.Printf("\n🎉 任务完成！视频文件： %s\n", finalPath)
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
		// 统一分辨率 1080x1920，防止素材比例不一报错
		filterComplex += fmt.Sprintf("[%d:v]scale=1080:1920:force_original_aspect_ratio=increase,crop=1080:1920,setsar=1,fps=30,format=yuv420p[v%d];", i, i)
		concatIn += fmt.Sprintf("[v%d]", i)
	}
	filterComplex += fmt.Sprintf("%sconcat=n=%d:v=1:a=0[vcat]", concatIn, len(videoInputs))

	args = append(args,
		"-filter_complex", filterComplex,
		"-map", "[vcat]", "-map", fmt.Sprintf("%d:a", audioIdx),
		"-c:v", "libx264", "-preset", "fast", "-c:a", "aac",
		"-shortest", // 核心：音频结束即切断画面
		"-y", outPath,
	)
	return exec.Command("ffmpeg", args...).Run()
}
