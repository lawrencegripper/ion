package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/lawrencegripper/ion/modules/helpers/Go/env"
	"github.com/lawrencegripper/ion/modules/helpers/Go/handler"
	"github.com/lawrencegripper/ion/modules/helpers/Go/log"
)

func main() {
	exitCode := 0
	defer func() { os.Exit(exitCode) }()

	env.MakeOutputDirs()

	log.Debug("Transcoding with FFMPEG")

	start := time.Now()

	files, err := ioutil.ReadDir(env.InputDataDir())
	if err != nil {
		exitCode = 13
		log.Info("Failed to read input directory")
		log.Info(err.Error())
		return
	}

	results := []string{}
	for _, inputFile := range files {
		transcodedFileName, err := transcode(filepath.Join(env.InputDataDir(), inputFile.Name()))
		if err != nil {
			exitCode = 14
			log.Info(fmt.Sprintf("Failed transcoding file: %s", inputFile.Name()))
			log.Info(err.Error())
			return
		}

		results = append(results, transcodedFileName)
	}

	elapsed := time.Since(start)

	handler.WriteInsights(handler.Insights{
		handler.Insight{
			Key:   "transcodeTimeSec",
			Value: fmt.Sprintf("%.6f", elapsed.Seconds()),
		},
	})

	handler.WriteEvents([]handler.Event{
		{
			Event: "file_transcoded",
			Files: results,
		},
	})
}

func transcode(inputFilePath string) (string, error) {
	_, inputFilename := filepath.Split(inputFilePath)
	outputFileName := fmt.Sprintf("%s-1280x720-h264.mp4", inputFilename)
	outputFilePath := filepath.Join(env.OutputDataDir(), outputFileName)
	useGPU := os.Getenv("FFMPEG_USE_GPU")
	var args string
	if useGPU == "false" {
		args = fmt.Sprintf(`-i %s -vcodec h264_nvenc %s`, inputFilePath, outputFilePath)
	} else {
		args = fmt.Sprintf(`-hwaccel cuvid -i %s -vf scale_npp=1280:720 -c:v h264_nvenc %s`, inputFilePath, outputFilePath)
	}
	log.Info("Running ffmpeg with args:")
	log.Info(args)
	consoleOutput, err := runFFMPEG(time.Duration(time.Second*300), false, strings.Split(args, " ")...)
	if err != nil {
		log.Fatal(consoleOutput)
		return "", err
	}
	log.Info(consoleOutput)
	return outputFileName, nil
}

func runFFMPEG(timeout time.Duration, isVerbose bool, args ...string) (string, error) {
	resultChan := make(chan string, 1)
	failedChan := make(chan error, 1)

	cmd := exec.Command("ffmpeg", args...)

	go func() {
		if isVerbose {
			fmt.Println("RunScript: Using verbose script output")
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout

			err := cmd.Run()
			if err != nil {
				failedChan <- err
				return
			}
			resultChan <- ""
		} else {
			output, err := cmd.CombinedOutput()
			resultChan <- string(output)
			if err != nil {
				log.Info(fmt.Sprintf("Failed running ffmpeg: %v", err))
			}
		}

	}()

	select {
	case err := <-failedChan:
		return "", err
	case output := <-resultChan:
		return string(output), nil
	case <-time.After(timeout):
		cmd.Process.Kill() //nolint: errcheck
		return "", fmt.Errorf("Timeout waiting for script after: %v", timeout)
	}
}
