package main

import (
    "os/exec"
)

func processVideoForFastStart(filePath string) (string, error) {
    outputFile := filePath + ".processing" 

    cmd := exec.Command(
        "ffmpeg",
        "-i", 
        filePath,
        "-c",
        "copy",
        "-movflags",
        "faststart",
        "-f",
        "mp4",
        outputFile,
    )

    err := cmd.Run()
    if err != nil {return "", nil}

    return outputFile, nil
}
