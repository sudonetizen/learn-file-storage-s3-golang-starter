package main

import (
    "bytes"
    "os/exec"
    "encoding/json"
)

func getVideoAspectRatio(filePath string) (string, error) {
    // configuring command 
    cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
    var b bytes.Buffer 
    cmd.Stdout = &b

    // running command
    err := cmd.Run()
    if err != nil {return "", err}

    // unmarshalling 
    var ss ArrayStreams 
    err = json.Unmarshal(b.Bytes(), &ss)

    // getting width and height
    width := ss.Streams[0].Width
    height := ss.Streams[0].Height

    // response 
    if width == 16*height/9  {return "16:9", nil}
    if height == 16*width/9 {return "9:16", nil}
    
    return "other", nil
}
