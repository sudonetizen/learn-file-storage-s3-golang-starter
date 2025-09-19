package main

import (
    "io"
    "os"
    "fmt"
    "mime"
	"net/http"
    "crypto/rand"
    "encoding/base64"
    "github.com/google/uuid"
    "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
    const uploadLimit = 10 << 30 
    r.Body = http.MaxBytesReader(w, r.Body, uploadLimit)

    // get video id 
    videoIdString := r.PathValue("videoID")
    videoId, err := uuid.Parse(videoIdString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

    // getting token 
    token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

    // getting user id from jwt 
    userId, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

    // getting video metadata
    video, err := cfg.db.GetVideo(videoId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "not found video with given videoID", err)
		return
	}

    // comparing user id from jwt with user id from video metadata 
    if userId != video.UserID {
        respondWithError(w, http.StatusUnauthorized, "not authorized", err)
        return
    }

    // getting video 
    file, _, err := r.FormFile("video")
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
        return
    }
    defer file.Close()

    // checking video format
    _, _, err = mime.ParseMediaType("video/mp4")
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "video format is not mp4", err)
        return
    }

    // creating temp file 
    fileTemp, err := os.CreateTemp("", "tubely-upload.mp4")
    if err != nil {
        respondWithError(w, 500, "error with creating temp file", err)
        return
    }
    defer os.Remove(fileTemp.Name())
    defer fileTemp.Close()

    // copying content from file to fileTemp 
    _, err = io.Copy(fileTemp, file)
    if err != nil {
        respondWithError(w, 500, "error with copying", err)
        return
    }
   
    // resetting fileTemp pointer to beginning 
    _, err = fileTemp.Seek(0, io.SeekStart)
    if err != nil {
        respondWithError(w, 500, "error with resetting pointer", err)
        return
    }
    
    // checking aspect ratio 
    aspectRatio, err := getVideoAspectRatio(fileTemp.Name())
    if err != nil {
        respondWithError(w, 500, "error with aspect ratio", err)
        return
    }

    folderS3 := ""
    if aspectRatio == "16:9" {
        folderS3 = "landscape/"
    } else if aspectRatio == "9:16" {
        folderS3 = "portrait/"
    } else {
        folderS3 = "other/"
    }

    // processing video faster opt 
    processedVideo, err := processVideoForFastStart(fileTemp.Name())
    if err != nil {
        respondWithError(w, 500, "error with processing video", err)
        return
    }

    filePV, err := os.Open(processedVideo)
    if err != nil {
        respondWithError(w, 500, "error with opening file", err)
        return
    }
    defer filePV.Close()

    // putting object into S3
    key := make([]byte, 32)
    rand.Read(key) 
    keyString := base64.RawURLEncoding.EncodeToString(key)
    keyString += ".mp4"

    fullKey := folderS3 + keyString

    videoType := "video/mp4"
    _, err = cfg.s3Client.PutObject(
        r.Context(),
        &s3.PutObjectInput{
            Bucket: &cfg.s3Bucket,
            Key: &fullKey,
            Body: filePV,
            ContentType: &videoType,
        },
    )

    if err != nil {
        respondWithError(w, 500, "error with putting object into s3", err)
        return
    }

    // updating video 
    dataUrl := fmt.Sprintf("https://%v.s3.%v.amazonaws.com/%v", cfg.s3Bucket, cfg.s3Region, fullKey)
    video.VideoURL= &dataUrl
    
    err = cfg.db.UpdateVideo(video)
    if err != nil {
        respondWithError(w, 500, "error with updating", err)
        return
    }

    // response
	respondWithJSON(w, http.StatusOK, video)
}
