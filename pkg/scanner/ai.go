package scanner

import (
	"os"

	"ineedspace/pkg/model"
)

// ScanAI scans for AI and LLM models.
func ScanAI() []model.Component {
	var components []model.Component

	// 1. Ollama Models
	ollamaPath := ExpandPath("~/.ollama/models")
	if DirExists(ollamaPath) {
		sz, count, _ := DirSize(ollamaPath)
		if sz > 10*1024*1024 { // Only report if > 10MB
			components = append(components, model.Component{
				ID:          "ai-ollama",
				Name:        "Ollama Local Models",
				Category:    model.CategoryAI,
				Description: "Local LLM weights stored in ~/.ollama/models",
				Path:        ollamaPath,
				SizeBytes:   sz,
				ItemCount:   count,
				Risk:        model.RiskNotice,
				Impact:      "Ollama models will be deleted. You can re-pull them with 'ollama run <model>'.",
				CleanAction: "Remove Ollama model blobs and manifests",
				Cleaner: func(dryRun bool) (int64, error) {
					if dryRun {
						return sz, nil
					}
					// If ollama CLI exists, try to delete models cleanly or remove directory
					_ = os.RemoveAll(ollamaPath)
					_ = os.MkdirAll(ollamaPath, 0755)
					return sz, nil
				},
			})
		}
	}

	// 2. LM Studio Models
	lmStudioPath := ExpandPath("~/.lmstudio/models")
	if DirExists(lmStudioPath) {
		sz, count, _ := DirSize(lmStudioPath)
		if sz > 10*1024*1024 {
			components = append(components, model.Component{
				ID:          "ai-lmstudio",
				Name:        "LM Studio Model Weights",
				Category:    model.CategoryAI,
				Description: "GGUF and Safetensors weights stored in ~/.lmstudio/models",
				Path:        lmStudioPath,
				SizeBytes:   sz,
				ItemCount:   count,
				Risk:        model.RiskNotice,
				Impact:      "LM Studio models will be removed. You can re-download them inside LM Studio.",
				CleanAction: "Delete downloaded models in ~/.lmstudio/models",
				Cleaner: func(dryRun bool) (int64, error) {
					if dryRun {
						return sz, nil
					}
					_ = os.RemoveAll(lmStudioPath)
					_ = os.MkdirAll(lmStudioPath, 0755)
					return sz, nil
				},
			})
		}
	}

	// 3. Whisper Speech Models Cache
	whisperPath := ExpandPath("~/.cache/whisper")
	if DirExists(whisperPath) {
		sz, count, _ := DirSize(whisperPath)
		if sz > 10*1024*1024 {
			components = append(components, model.Component{
				ID:          "ai-whisper",
				Name:        "OpenAI Whisper Audio Models",
				Category:    model.CategoryAI,
				Description: "Cached Whisper speech-to-text models in ~/.cache/whisper",
				Path:        whisperPath,
				SizeBytes:   sz,
				ItemCount:   count,
				Risk:        model.RiskSafe,
				Impact:      "Cached Whisper models will be wiped; re-downloaded automatically when needed.",
				CleanAction: "Delete ~/.cache/whisper",
				Cleaner: func(dryRun bool) (int64, error) {
					if dryRun {
						return sz, nil
					}
					return sz, os.RemoveAll(whisperPath)
				},
			})
		}
	}

	// 4. HuggingFace Hub Cache
	hfPath := ExpandPath("~/.cache/huggingface")
	if DirExists(hfPath) {
		sz, count, _ := DirSize(hfPath)
		if sz > 10*1024*1024 {
			components = append(components, model.Component{
				ID:          "ai-huggingface",
				Name:        "HuggingFace Hub Cache",
				Category:    model.CategoryAI,
				Description: "Downloaded datasets, tokenizers and models from HuggingFace",
				Path:        hfPath,
				SizeBytes:   sz,
				ItemCount:   count,
				Risk:        model.RiskNotice,
				Impact:      "Cached weights and tokenizers will be deleted. Will re-download on next python script run.",
				CleanAction: "Delete ~/.cache/huggingface",
				Cleaner: func(dryRun bool) (int64, error) {
					if dryRun {
						return sz, nil
					}
					return sz, os.RemoveAll(hfPath)
				},
			})
		}
	}

	return components
}
