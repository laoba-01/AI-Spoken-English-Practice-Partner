// internal/model/package/pkg/oss/aliyun_oss.go
package oss

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type AliyunOSS struct {
	client *oss.Client
	bucket *oss.Bucket
}

func NewAliyunOSS() *AliyunOSS {
	endpoint := os.Getenv("ALIYUN_OSS_ENDPOINT")
	bucketName := os.Getenv("ALIYUN_OSS_BUCKET")
	accessKeyID := os.Getenv("ALIYUN_ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("ALIYUN_ACCESS_KEY_SECRET")

	if endpoint == "" || bucketName == "" || accessKeyID == "" {
		log.Println("⚠ 未配置 OSS (ALIYUN_OSS_ENDPOINT/ALIYUN_OSS_BUCKET)，音频文件存储在本地")
		return nil
	}

	client, err := oss.New(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		log.Println("⚠ OSS 客户端创建失败，音频文件存储在本地:", err)
		return nil
	}

	bucket, err := client.Bucket(bucketName)
	if err != nil {
		log.Println("⚠ OSS Bucket 获取失败，音频文件存储在本地:", err)
		return nil
	}

	log.Println("✓ OSS 已配置 (bucket=" + bucketName + ")")
	return &AliyunOSS{client: client, bucket: bucket}
}

func (o *AliyunOSS) UploadMP3(data []byte, filename string) (string, error) {
	key := "audio/" + filename
	err := o.bucket.PutObject(key, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("OSS上传失败: %w", err)
	}

	// 如果有自定义域名则用，否则用默认 OSS 域名
	domain := os.Getenv("ALIYUN_OSS_DOMAIN")
	if domain == "" {
		endpoint := os.Getenv("ALIYUN_OSS_ENDPOINT")
		bucketName := os.Getenv("ALIYUN_OSS_BUCKET")
		domain = fmt.Sprintf("https://%s.%s", bucketName, endpoint)
	}

	return fmt.Sprintf("%s/%s", domain, key), nil
}
