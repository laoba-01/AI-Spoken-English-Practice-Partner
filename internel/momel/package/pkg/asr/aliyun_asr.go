package asr

import (
	"context"
	"os"
	"time"

	"github.com/aliyun/alibabacloud-nls-go-sdk/nls"
)

type AliyunASR struct {
	accessKeyID     string
	accessKeySecret string
	appKey          string
}

func NewAliyunASR() *AliyunASR {
	return &AliyunASR{
		accessKeyID:     os.Getenv("ALIYUN_ACCESS_KEY_ID"),
		accessKeySecret: os.Getenv("ALIYUN_ACCESS_KEY_SECRET"),
		appKey:          os.Getenv("ALIYUN_ASR_APPKEY"),
	}
}

// RecognizeMP3 识别MP3语音文件（支持60秒以内）
func (a *AliyunASR) RecognizeMP3(mp3Data []byte) (string, error) {
	client, err := nls.NewClient(a.accessKeyID, a.accessKeySecret)
	if err != nil {
		return "", err
	}
	defer client.Close()

	req := nls.NewSpeechRecognizerRequest(a.appKey)
	req.SetFormat("mp3")
	req.SetSampleRate(16000)
	req.SetEnablePunctuation(true)
	req.SetEnableITN(true)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.SpeechRecognizer(ctx, req, mp3Data)
	if err != nil {
		return "", err
	}

	return resp.Result.Text, nil
}