package asr

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"

	nls "github.com/aliyun/alibabacloud-nls-go-sdk"
)

type AliyunASR struct {
	accessKeyID     string
	accessKeySecret string
	appKey          string
	url             string
}

func NewAliyunASR() *AliyunASR {
	url := os.Getenv("ALIYUN_NLS_URL")
	if url == "" {
		url = nls.DEFAULT_URL
	}
	return &AliyunASR{
		accessKeyID:     os.Getenv("ALIYUN_ACCESS_KEY_ID"),
		accessKeySecret: os.Getenv("ALIYUN_ACCESS_KEY_SECRET"),
		appKey:          os.Getenv("ALIYUN_ASR_APPKEY"),
		url:             url,
	}
}

// ASRResponse 阿里云ASR识别响应JSON结构
type ASRResponse struct {
	Header struct {
		Name string `json:"name"`
	} `json:"header"`
	Payload struct {
		Result string `json:"result"`
	} `json:"payload"`
}

// RecognizeWebM 识别WebM语音（浏览器原生录音格式）
func (a *AliyunASR) RecognizeWebM(webmData []byte) (string, error) {
	return a.recognize(webmData, "webm")
}

// RecognizeMP3 识别MP3语音文件
func (a *AliyunASR) RecognizeMP3(mp3Data []byte) (string, error) {
	return a.recognize(mp3Data, "mp3")
}

func (a *AliyunASR) recognize(data []byte, format string) (string, error) {
	if a.accessKeyID == "" || a.accessKeySecret == "" || a.appKey == "" {
		return "", errors.New("阿里云ASR未配置: 请设置 ALIYUN_ACCESS_KEY_ID, ALIYUN_ACCESS_KEY_SECRET, ALIYUN_ASR_APPKEY")
	}

	config, err := nls.NewConnectionConfigWithAKInfoDefault(
		a.url, a.appKey, a.accessKeyID, a.accessKeySecret,
	)
	if err != nil {
		return "", err
	}

	var resultText string
	var resultErr error
	var doneOnce sync.Once
	done := make(chan struct{})
	closeDone := func() { doneOnce.Do(func() { close(done) }) }

	sr, err := nls.NewSpeechRecognition(
		config,
		nls.DefaultNlsLog(),
		// onTaskFailed
		func(text string, param interface{}) {
			resultErr = errors.New("ASR task failed: " + text)
			closeDone()
		},
		// onStarted
		func(text string, param interface{}) {},
		// onResultChanged
		func(text string, param interface{}) {
			var resp ASRResponse
			if err := json.Unmarshal([]byte(text), &resp); err == nil {
				if resp.Payload.Result != "" {
					resultText = resp.Payload.Result
				}
			}
		},
		// onCompleted
		func(text string, param interface{}) {
			// onCompleted 中也包含最终识别结果（一句话识别场景下，
			// 结果通常通过 onCompleted 而非 onResultChanged 返回）
			var resp ASRResponse
			if err := json.Unmarshal([]byte(text), &resp); err == nil {
				if resp.Payload.Result != "" {
					resultText = resp.Payload.Result
				}
			}
			closeDone()
		},
		// onClosed
		func(param interface{}) {},
		nil, // userParam
	)
	if err != nil {
		return "", err
	}

	param := nls.DefaultSpeechRecognitionParam()
	param.Format = format
	param.SampleRate = 16000
	param.EnableIntermediateResult = false
	param.EnablePunctuationPrediction = true
	param.EnableInverseTextNormalization = true

	ready, err := sr.Start(param, nil)
	if err != nil {
		sr.Shutdown()
		return "", err
	}
	<-ready

	if err := sr.SendAudioData(data); err != nil {
		sr.Shutdown()
		return "", err
	}

	stopReady, err := sr.Stop()
	if err != nil {
		sr.Shutdown()
		return "", err
	}
	<-stopReady

	select {
	case <-done:
	case <-time.After(15 * time.Second):
		resultErr = errors.New("ASR识别超时（15秒）")
		closeDone()
	}

	sr.Shutdown()

	if resultErr != nil {
		return "", resultErr
	}
	return resultText, nil
}
