package tts

import (
	"context"
	"os"
	"time"

	"github.com/aliyun/alibabacloud-nls-go-sdk/nls"
)

type AliyunTTS struct {
	accessKeyID     string
	accessKeySecret string
	appKey          string
}

func NewAliyunTTS() *AliyunTTS {
	return &AliyunTTS{
		accessKeyID:     os.Getenv("ALIYUN_ACCESS_KEY_ID"),
		accessKeySecret: os.Getenv("ALIYUN_ACCESS_KEY_SECRET"),
		appKey:          os.Getenv("ALIYUN_TTS_APPKEY"),
	}
}

// TextToMP3 文字转MP3语音
func (t *AliyunTTS) TextToMP3(text string) ([]byte, error) {
	client, err := nls.NewClient(t.accessKeyID, t.accessKeySecret)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	req := nls.NewSpeechSynthesizerRequest(t.appKey)
	req.SetText(text)
	req.SetVoice("en-US-JennyNeural") // 英文女声，最自然
	req.SetFormat("mp3")
	req.SetSampleRate(16000)
	req.SetSpeechRate(0)
	req.SetPitchRate(0)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.SpeechSynthesizer(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Audio, nil
}