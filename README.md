# AutoClip-Go 🎬

![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-green.svg)
![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)

**AutoClip-Go** 是一个基于 Go 语言开发的全自动短视频生成流水线。它利用 AI 模型（DeepSeek）进行内容策划，结合 Pexels 素材库和边缘 TTS 技术，通过一行命令即可将“一个想法”转化为“一支完成度极高的短视频”。

> 🚀 **项目初衷**：探索 Go 语言在多媒体处理与 AI Agent 编排领域的应用，实现从文案到视频的自动化闭环。

## ✨ 核心特性 (Key Features)

* **🧠 AI 驱动策划**：集成 **DeepSeek-V3** 模型，自动生成快节奏口播文案及匹配的英文搜索关键词，内置鲁棒的 JSON 提取算法，防止 AI 输出格式错误。
* **⚡ 高并发素材下载**：
    * 使用 `sync.WaitGroup` 和 **Semaphore (信号量)** 模式控制并发度（默认 3 线程），在极大化下载速度的同时防止触发 API 速率限制。
    * 内置智能重试与兜底策略，当关键词搜索无果时自动回退算法。
* **🎥 工业级视频合成**：
    * 通过 `os/exec` 编排 **FFmpeg** 进行复杂的滤镜处理（Scale, Crop, Setsar）。
    * 实现**音画对其** (`-shortest`)，确保视频画面随音频结束精准截断。
* **🎙️ 边缘 TTS 生成**：调用 `edge-tts` 生成高质量、自然的中文语音（zh-CN-YunxiNeural）。
* **🧹 自动化运维**：执行完毕后自动粉碎中间临时文件，保持工作区整洁。

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
