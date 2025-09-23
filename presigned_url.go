package main

import (
    "time"
    "context"
    "strings"
    "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
    presignedClient := s3.NewPresignClient(s3Client)

    presignedRequest, err := presignedClient.PresignGetObject(
        context.Background(), 
        &s3.GetObjectInput{Bucket: &bucket, Key: &key},
    )

    if err != nil {return "", err}

    return presignedRequest.URL, nil
}

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
    s := video.VideoURL
    bucketAndKey := strings.Split(*s, ",")
    bucket := bucketAndKey[0]
    key := bucketAndKey[1]
    
    presignedURL, err := generatePresignedURL(cfg.s3Client, bucket, key, 60 * time.Minute)
    if err != nil {return video, err}

    video.VideoURL = &presignedURL

    return video, nil
}
