package tts

import (
	"errors"
	"os"
	"sync"
	"time"

	nls "github.com/aliyun/alibabacloud-nls-go-sdk"
)

type AliyunTTS struct {
	accessKeyID     string
	accessKeySecret string
	appKey          string
	url             string
}

func NewAliyunTTS() *AliyunTTS {
	url := os.Getenv("ALIYUN_NLS_URL")
	if url == "" {
		url = nls.DEFAULT_URL
	}
	return &AliyunTTS{
		accessKeyID:     os.Getenv("ALIYUN_ACCESS_KEY_ID"),
		accessKeySecret: os.Getenv("ALIYUN_ACCESS_KEY_SECRET"),
		appKey:          os.Getenv("ALIYUN_TTS_APPKEY"),
		url:             url,
	}
}

// TextToMP3 文字转MP3语音
func (t *AliyunTTS) TextToMP3(text string) ([]byte, error) {
	if t.accessKeyID == "" || t.accessKeySecret == "" || t.appKey == "" {
		return nil, errors.New("阿里云TTS未配置: 请设置 ALIYUN_ACCESS_KEY_ID, ALIYUN_ACCESS_KEY_SECRET, ALIYUN_TTS_APPKEY")
	}

	config, err := nls.NewConnectionConfigWithAKInfoDefault(
		t.url, t.appKey, t.accessKeyID, t.accessKeySecret,
	)
	if err != nil {
		return nil, err
	}

	var audioData []byte
	var resultErr error
	var doneOnce sync.Once
	done := make(chan struct{})
	closeDone := func() { doneOnce.Do(func() { close(done) }) }

	tts, err := nls.NewSpeechSynthesis(
		config,
		nls.DefaultNlsLog(),
		false, // realtimeLongText = false，普通文本合成
		// onTaskFailed
		func(text string, param interface{}) {
			resultErr = errors.New("TTS task failed: " + text)
			closeDone()
		},
		// onSynthesisResult — 接收合成的音频数据
		func(data []byte, param interface{}) {
			audioData = append(audioData, data...)
		},
		// onMetaInfo
		func(text string, param interface{}) {},
		// onCompleted
		func(text string, param interface{}) {
			closeDone()
		},
		// onClosed
		func(param interface{}) {},
		nil, // userParam
	)
	if err != nil {
		return nil, err
	}

	param := nls.DefaultSpeechSynthesisParam()
	param.Voice = "Abby" // 美音女声
	param.Format = "mp3"
	param.SampleRate = 16000
	param.Volume = 50
	param.SpeechRate = 0
	param.PitchRate = 0

	ready, err := tts.Start(text, param, nil)
	if err != nil {
		tts.Shutdown()
		return nil, err
	}
	<-ready

	select {
	case <-done:
	case <-time.After(15 * time.Second):
		resultErr = errors.New("TTS合成超时（15秒）")
		closeDone()
	}

	tts.Shutdown()

	if resultErr != nil {
		return nil, resultErr
	}
	return audioData, nil
}
