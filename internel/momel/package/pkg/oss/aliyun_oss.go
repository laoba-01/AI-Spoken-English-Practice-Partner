// internal/model/package/pkg/oss/aliyun_oss.go
package oss

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type AliyunOSS struct {
	client    *oss.Client
	bucket    *oss.Bucket
	cdnDomain string // CDN 加速域名（如 https://cdn.example.com）
}

func NewAliyunOSS() *AliyunOSS {
	endpoint := os.Getenv("ALIYUN_OSS_ENDPOINT")
	bucketName := os.Getenv("ALIYUN_OSS_BUCKET")
	accessKeyID := os.Getenv("ALIYUN_ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("ALIYUN_ACCESS_KEY_SECRET")
	cdnDomain := os.Getenv("ALIYUN_OSS_DOMAIN") // CDN 加速域名

	if endpoint == "" || bucketName == "" || accessKeyID == "" {
		log.Println("⚠ 未配置 OSS (ALIYUN_OSS_ENDPOINT/ALIYUN_OSS_BUCKET)，音频文件存储在本地")
		return nil
	}

	client, err := oss.New("https://"+endpoint, accessKeyID, accessKeySecret)
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
	if cdnDomain != "" {
		log.Println("✓ CDN 加速已启用 (domain=" + cdnDomain + ")")
	}
	return &AliyunOSS{client: client, bucket: bucket, cdnDomain: cdnDomain}
}

func (o *AliyunOSS) UploadMP3(data []byte, filename string) (string, error) {
	key := "audio/" + filename
	err := o.bucket.PutObject(key, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("OSS上传失败: %w", err)
	}

	// CDN 加速域名优先（速度快，无需签名）
	if o.cdnDomain != "" {
		return fmt.Sprintf("%s/%s", o.cdnDomain, key), nil
	}

	// 降级：生成签名 URL（1小时有效），私有 bucket 也能访问
	signedURL, err := o.bucket.SignURL(key, oss.HTTPGet, int64((1 * time.Hour).Seconds()))
	if err != nil {
		return "", fmt.Errorf("OSS签名URL生成失败: %w", err)
	}

	return signedURL, nil
}
