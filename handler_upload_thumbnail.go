package main

import (
    "io"
    "os"
	"fmt"
    "strings"
	"net/http"
    "path/filepath"

	"github.com/google/uuid"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}


	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
    const maxMemory = 10 << 20 
    r.ParseMultipartForm(maxMemory)

    // getting thumbnail
    file, header, err := r.FormFile("thumbnail")
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
        return
    }
    defer file.Close()

    // content type 
    contentType := header.Header.Get("Content-Type")
    fmt.Println(contentType)

    // getting video metadata 
    vde, err := cfg.db.GetVideo(videoID)
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "Unable to find video", err)
        return
    }

    // checking userID with userID at video's metadata
    if userID != vde.UserID {
        respondWithError(w, http.StatusUnauthorized, "Unable to find video", err)
        return
    }

    // saving as a file 
    fileFormat := strings.ReplaceAll(contentType, "image/", "")
    fileFormat = strings.TrimSpace(fileFormat)
    path := filepath.Join("assets", videoIDString + "." + fileFormat)

    fle, err := os.Create(path)
    if err != nil {
        respondWithError(w, 500, "error with creating path", err)
        return
    }
    defer fle.Close()
    
    _, err = io.Copy(fle, file)
    if err != nil {
        respondWithError(w, 500, "error with copying file", err)
        return
    }
  
    // updating video 
    dataURL := fmt.Sprintf("http://localhost:%v/assets/%v.%v", cfg.port, videoIDString, fileFormat)
    vde.ThumbnailURL = &dataURL

    err = cfg.db.UpdateVideo(vde)
    if err != nil {
        respondWithError(w, 500, "error with converting", err)
        return
    }

	respondWithJSON(w, http.StatusOK, vde)
}
