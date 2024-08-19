package main

import (
	"fmt"
	"image"
	"log"
	"math/rand/v2"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	imageprocessing "pipeline/src/image_processing"
)

var (
	iterations = 5
	imagesDir  = "./images"
)

type IImagePipeline interface {
	Run()
}

type Job struct {
	Image      image.Image
	OutputPath string
}
type ImagePipeline struct {
	filePath []string
}

func NewImagePipeline(filePath []string) *ImagePipeline {
	return &ImagePipeline{
		filePath: filePath,
	}
}

func (ip *ImagePipeline) Run() {
	var wg sync.WaitGroup
	for _, path := range ip.filePath {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()

			chan1 := ip.loadImages(p)
			chan2 := ip.resize(chan1)
			chan3 := ip.grayscale(chan2)
			isOK := ip.saveImage(chan3)

			if !isOK {
				fmt.Printf("Error image: %v\n", p)
			}

		}(path)
	}

	wg.Wait()

}

func (ip *ImagePipeline) loadImages(path string) <-chan Job {
	out := make(chan Job)
	go func() {
		defer close(out)

		job := Job{
			OutputPath: strings.Replace(path, "images/", "images/output/", 1),
		}

		job.Image = imageprocessing.ReadImage(path)
		out <- job

	}()

	return out
}

func (ip *ImagePipeline) resize(job <-chan Job) <-chan Job {
	out := make(chan Job)

	go func() {
		defer close(out)

		for j := range job {

			resizedImg := imageprocessing.Resize(j.Image)
			j.Image = resizedImg

			out <- j

		}
	}()

	return out
}

func (ip *ImagePipeline) grayscale(job <-chan Job) <-chan Job {
	out := make(chan Job)

	go func() {
		defer close(out)

		for j := range job {

			grayImg := imageprocessing.Grayscale(j.Image)
			j.Image = grayImg

			out <- j
		}
	}()

	return out
}

func (ip *ImagePipeline) saveImage(input <-chan Job) bool {

	for job := range input {
		err := imageprocessing.WriteImage(job.OutputPath, job.Image)
		if err != nil {
			fmt.Println()
			return false
		}
	}
	return true
}

type NoPipeline struct {
	filePath []string
}

func NewNoPipeline(filePath []string) *NoPipeline {
	return &NoPipeline{
		filePath: filePath,
	}
}

func (np *NoPipeline) Run() {
	for _, path := range np.filePath {

		image := imageprocessing.ReadImage(path)
		grayImg := imageprocessing.Grayscale(image)
		resizedImg := imageprocessing.Resize(grayImg)
		outputPath := strings.Replace(path, "images/", "images/output2/", 1)
		imageprocessing.WriteImage(outputPath, resizedImg)

	}
}

type NoPipelineWithGoRoutines struct {
	filePath []string
}

func NewNoPipelineWithGoRoutines(filePath []string) *NoPipelineWithGoRoutines {

	return &NoPipelineWithGoRoutines{
		filePath: filePath,
	}
}

func (npgr *NoPipelineWithGoRoutines) Run() {
	var wg sync.WaitGroup
	for _, path := range npgr.filePath {
		wg.Add(1)
		go func(p string) {

			defer wg.Done()

			image := imageprocessing.ReadImage(p)
			grayImg := imageprocessing.Grayscale(image)
			resizedImg := imageprocessing.Resize(grayImg)
			outputPath := strings.Replace(p, "images/", "images/output3/", 1)
			imageprocessing.WriteImage(outputPath, resizedImg)
		}(path)

	}
	wg.Wait()
}

func Benchmark(imagesFilePaths []string) {

	imgPipeline := NewImagePipeline(imagesFilePaths)
	noPipeline := NewNoPipeline(imagesFilePaths)
	noPipelineWithGoRoutines := NewNoPipelineWithGoRoutines(imagesFilePaths)

	tests := []struct {
		name string
		fn   IImagePipeline
	}{
		{"Pipeline", imgPipeline},
		{"Sequential Processing", noPipeline},
		{"Parallel Processing without Pipeline", noPipelineWithGoRoutines},
	}

	results := make(map[string][]time.Duration)

	for i := 0; i < iterations; i++ {
		fmt.Printf("Iteration %d:\n", i+1)
		rand.Shuffle(len(tests), func(i, j int) { tests[i], tests[j] = tests[j], tests[i] })

		for _, test := range tests {
			os.RemoveAll("./images/output")
			os.RemoveAll("./images/output2")
			os.RemoveAll("./images/output3")
			runtime.GC() // Force garbage collection before each test

			start := time.Now()
			test.fn.Run()
			duration := time.Since(start)

			results[test.name] = append(results[test.name], duration)
			fmt.Printf("%s: %v\n", test.name, duration)
		}
		fmt.Println()
	}

	fmt.Println("Average Results:")
	for name, durations := range results {
		var total time.Duration
		for _, d := range durations {
			total += d
		}
		avg := total / time.Duration(len(durations))
		fmt.Printf("%s: %v\n", name, avg)
	}
	os.RemoveAll("./images/output")
	os.RemoveAll("./images/output2")
	os.RemoveAll("./images/output3")
}

func main() {

	files, err := os.ReadDir(imagesDir)

	if err != nil {
		log.Panic("List files Error", err)
		return
	}
	imagesFilePaths := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {

			filePath := filepath.Join(imagesDir, file.Name())
			imagesFilePaths = append(imagesFilePaths, filePath)

		}

	}

	Benchmark(imagesFilePaths)

}
