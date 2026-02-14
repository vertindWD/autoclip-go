# AutoClip-Go 🎬

![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)

**AutoClip-Go** 是一个基于 Go 语言 (v1.24) 开发的全自动短视频生产流水线。

本项目利用 **Go 语言的高并发特性** 编排 AI Agent，通过集成 DeepSeek (文案策划)、Pexels (高清素材) 和 Edge-TTS (语音合成)，实现从“一个主题想法”到“一支完整短视频”的自动化闭环。

## ✨ 核心特性 (Features)

* **🧠 AI 驱动策划**: 
    * 集成 **DeepSeek-V3** 模型，自动生成快节奏口播文案及对应的英文搜索关键词。
    * 内置鲁棒的 JSON 提取与清洗算法，有效处理 LLM 输出的非结构化数据。
* **⚡ 高并发素材下载**: 
    * 使用 `sync.WaitGroup` 和 **Semaphore (信号量)** 模式控制并发度（默认 3 线程），最大化利用带宽同时避免 Pexels API 限流。
    * 智能重试与回退机制：确保素材抓取的稳定性。
* **🎥 工业级视频合成**: 
    * 通过 `os/exec` 编排 **FFmpeg** 进行复杂的滤镜链处理（Scale, Crop, Setsar），强制统一素材为 1080x1920 (9:16) 竖屏格式。
    * 实现**音画精准对齐** (`-shortest`)，确保视频画面随音频结束自动截断。
* **🎙️ 边缘 TTS 生成**: 
    * 调用 `edge-tts` 生成高质量、自然的中文语音（zh-CN-YunxiNeural）。
* **🧹 自动化运维**: 
    * 任务完成后自动粉碎中间临时素材，保持工作目录整洁。

## 🏗️ 架构流程

```mermaid
graph LR
    A[用户输入主题] --> B(DeepSeek AI 策划)
    B --> C{解析 JSON}
    C -->|Script| D[Edge-TTS 生成音频]
    C -->|Keywords| E[并发下载 Pexels 素材]
    E --> F[FFmpeg 渲染合成]
    D --> F
    F --> G[输出: final_movie.mp4]

## 🛠️ 快速开始 (Quick Start)

### 1. 环境准备
确保你的系统已安装以下依赖：
* **Go** (1.24+)
* **Python 3** (用于运行 edge-tts)
* **FFmpeg** (需加入系统 PATH 环境变量)

安装 Python 依赖：

    pip install edge-tts

### 2. 安装项目

    git clone https://github.com/vertindWD/autoclip-go.git
    cd autoclip-go
    go mod tidy

### 3. 配置密钥
在.env文件填入你的 API Key：

    # .env
    DEEPSEEK_API_KEY=your_deepseek_key_here
    PEXELS_API_KEY=your_pexels_key_here

    # 可选配置（如果不填则使用默认值）
    SCRIPT_WORD_COUNT=300
    VIDEO_CLIP_COUNT=8

### 4. 运行生成
直接运行主程序，通过 `-t` 参数传入视频主题：

    # 示例：生成关于“沉没成本”的科普视频
    go run . -t "沉没成本的陷阱"

程序运行结束后，成品视频将保存在 `./output/final_movie.mp4`。

## 📂 项目结构

    .
    ├── main.go           # 主程序入口：编排整个流水线（AI -> 下载 -> 合成）
    ├── llm.go            # AI 模块：负责与 DeepSeek API 交互及 JSON 清洗
    ├── material.go       # 素材模块：负责 Pexels 视频搜索与并发下载
    ├── utils.go          # 工具模块：文件清理等辅助功能
    ├── go.mod            # Go 依赖定义
    └── .env              # 配置文件（请勿提交到 Git）

## ⚙️ 参数说明

| 参数 | 描述 | 默认值 |
| :--- | :--- | :--- |
| `-t` | 视频主题 (Topic) | `"沉没成本的陷阱"` |
| `SCRIPT_WORD_COUNT` | (.env) 文案字数要求 | `300` |
| `VIDEO_CLIP_COUNT` | (.env) 下载素材数量 | `8` |

## 📅 开发计划 (Roadmap)

- [x] 基础流程跑通 (AI -> 素材 -> 合成)
- [x] Go 协程并发下载控制
- [x] 鲁棒的 JSON 解析器
- [ ] 支持更多 LLM 模型 (OpenAI, Claude)


## 🤝 贡献 (Contributing)

欢迎提交 Issue 或 Pull Request！如果你对 **Go 并发编程**感兴趣，请随时联系我。

## 📄 许可证 (License)

本项目基于 [MIT License](LICENSE) 开源。

---
*Created by [Zhao](https://github.com/vertindWD)*
