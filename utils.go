package main

import "os"

// CleanTempFiles 只清理素材，不碰 final_movie.mp4
func CleanTempFiles(videoClips []string, audioFile string) {
	for _, f := range videoClips {
		_ = os.Remove(f)
	}
	_ = os.Remove(audioFile)
}
