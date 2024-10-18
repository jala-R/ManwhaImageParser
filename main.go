package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"gocv.io/x/gocv"
)

const configFilePath = "./config.json"

var ptr = 0

func getLaplacianScore(mat *gocv.Mat) float64 {

	lap := gocv.NewMat()
	defer lap.Close()
	gocv.Laplacian(*mat, &lap, gocv.MatTypeCV64F, 1, 1, 0, gocv.BorderDefault)

	mat1 := gocv.NewMat()
	mat2 := gocv.NewMat()

	gocv.MeanStdDev(lap, &mat1, &mat2)

	variance := (mat2.GetDoubleAt(0, 0))

	temp := variance * variance

	return temp

}

// func main() {
// 	path := "./Output1"
// 	var paths []string
// 	filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
// 		if info.IsDir() {
// 			return nil
// 		}
// 		paths = append(paths, path)
// 		// os.Exit(1)
// 		return nil
// 	})

// 	sort.Slice(paths, func(i, j int) bool {
// 		path1 := strings.Split(strings.Split(paths[i], "/")[1], ".")[0]
// 		path2 := strings.Split(strings.Split(paths[j], "/")[1], ".")[0]

// 		num1, _ := strconv.ParseInt(path1, 10, 16)
// 		num2, _ := strconv.ParseInt(path2, 10, 16)

// 		return num1 < num2
// 	})

// 	for _, i := range paths {
// 		isBlur(i)
// 	}
// }

type ConfigFile struct {
	Stage  string `json:"Stage"`
	Output string `json:"Output"`
	Input  string `json:"Input"`
	Offset int    `json:"Offset"`
	Limit  int    `json:"limit"`
}

func test2(cfg ConfigFile) {
	dirs, err := os.ReadDir(cfg.Output)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var fileNames []string

	for i := range dirs {
		if !dirs[i].IsDir() {
			fileNames = append(fileNames, dirs[i].Name())
		}
	}

	sort.Slice(fileNames, func(i, j int) bool {
		f1 := strings.Split(fileNames[i], ".")[0]
		f2 := strings.Split(fileNames[j], ".")[0]

		file1, _ := strconv.ParseInt(f1, 10, 64)
		file2, _ := strconv.ParseInt(f2, 10, 64)

		return file1 < file2
	})

	for i := range fileNames {
		if i == 0 {
			continue
		}

		frame := gocv.NewMat()
		mat1 := gocv.IMRead(fmt.Sprintf("%s/%s", cfg.Output, fileNames[i]), gocv.IMReadGrayScale)
		mat2 := gocv.IMRead(fmt.Sprintf("%s/%s", cfg.Output, fileNames[i-1]), gocv.IMReadGrayScale)

		gocv.AbsDiff(mat1, mat2, &frame)

		// thres := gocv.NewMat()
		// gocv.Threshold(frame, &thres, 30, 255, gocv.ThresholdBinary)

		diff := gocv.CountNonZero(frame)

		percentage := getPercentage(float64(frame.Rows()*frame.Cols()), float64(diff))
		val := gocv.IMWrite(fmt.Sprintf("%s/%d.png", "./part2", i), frame)
		// fmt.Printf("wrote %d  -> %.2f%% \n", i, percentage)
		if percentage >= 65 {
			if val {
				fmt.Printf("wrote %d  -> %.2f%% \n", i, percentage)
			}
		}

		mat1.Close()
		mat2.Close()
		frame.Close()

		// fmt.Println(fileNames[i])
	}
}

func getPercentage(total, given float64) float64 {
	return (given / total) * 100
}

func configParser() ConfigFile {
	file, err := os.Open(configFilePath)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var cfg ConfigFile

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return cfg
}

func main() {
	fmt.Println("kjnbjfn")
	var cfg = configParser()
	fmt.Println(cfg)
	if cfg.Stage == "test" {
		experiment(cfg)
	} else if cfg.Stage == "prod" {
		prod(cfg)
	}
	test2(cfg)
}

func isImageSwitch(prev, current gocv.Mat) bool {
	if prev.Empty() {
		return false
	}
	mat1 := gocv.NewMat()
	mat2 := gocv.NewMat()
	frame := gocv.NewMat()

	defer func() {
		mat1.Close()
		mat2.Close()
		frame.Close()
	}()

	gocv.CvtColor(prev, &mat1, gocv.ColorBGRToGray)
	gocv.CvtColor(current, &mat2, gocv.ColorBGRToGray)

	gocv.AbsDiff(mat1, mat2, &frame)

	// thres := gocv.NewMat()
	// gocv.Threshold(frame, &thres, 30, 255, gocv.ThresholdBinary)

	diff := gocv.CountNonZero(frame)

	percentage := getPercentage(float64(frame.Rows()*frame.Cols()), float64(diff))
	return percentage >= 90
}

func isInvalidImage(lap float64) bool {
	return lap <= 17
}

func store(frame *gocv.Mat, cfg ConfigFile) bool {

	if frame == nil {
		return false
	}
	if frame.Empty() {
		return false
	}

	ptr++
	val := gocv.IMWrite(fmt.Sprintf("%s/%d.png", cfg.Output, ptr), *frame)

	return val
}

func prod(cfg ConfigFile) {

	fmt.Println("Running")
	vid, err := gocv.VideoCaptureFile(cfg.Input)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer vid.Close()

	fps := vid.Get(gocv.VideoCaptureFPS)

	frameCnt := vid.Get(gocv.VideoCaptureFrameCount)

	fmt.Printf("The video frame rate => %.4f \n", fps)
	fmt.Printf("The video frame counts => %.4f \n", frameCnt)

	frame := gocv.NewMat()
	defer frame.Close()

	prevFrame := gocv.NewMat()
	defer prevFrame.Close()

	maxLaplacianScore := math.SmallestNonzeroFloat64
	lastMaxFrame := gocv.NewMat()
	defer lastMaxFrame.Close()

	if cfg.Offset != 0 {
		vid.Grab(cfg.Offset)
	}

	for i := 0; true; i++ {
		// fmt.Println("fdfd")
		// frame.Close()
		if ok := vid.Read(&frame); !ok {
			fmt.Println("Frame Read completed")
			break
		}

		if i >= cfg.Limit {
			break
		}

		// if i < 40000 {
		// 	continue
		// }

		var laplacianScore = getLaplacianScore(&frame)

		fmt.Printf("\r Frame Cnt : %d  lscore %f", i+cfg.Offset, laplacianScore)

		if isInvalidImage(laplacianScore) || isImageSwitch(prevFrame, frame) {

			val := store(&lastMaxFrame, cfg)

			if val {
				fmt.Printf("Stored image %d  %.f\n ", ptr, getLaplacianScore(&lastMaxFrame))
			}

			maxLaplacianScore = math.SmallestNonzeroFloat64
			lastMaxFrame.Close()
			lastMaxFrame = gocv.NewMat()
		} else if maxLaplacianScore < laplacianScore && laplacianScore > 17 {
			maxLaplacianScore = laplacianScore
			lastMaxFrame.Close()
			lastMaxFrame = (frame.Clone())

		}
		prevFrame.Close()
		prevFrame = frame.Clone()
		// }

	}

	fmt.Println("Done...")

}

func experiment(cfg ConfigFile) {

	fmt.Println("Running")
	vid, err := gocv.VideoCaptureFile(cfg.Input)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	defer vid.Close()

	fps := vid.Get(gocv.VideoCaptureFPS)

	frameCnt := vid.Get(gocv.VideoCaptureFrameCount)

	fmt.Printf("The video frame rate => %.4f \n", fps)
	fmt.Printf("The video frame counts => %.4f \n", frameCnt)

	frame := gocv.NewMat()
	defer frame.Close()

	// maxLaplacianScore := math.SmallestNonzeroFloat64
	// var lastMaxFrame gocv.Mat

	// if cfg.Offset != 0 {
	vid.Grab(cfg.Offset)
	// }

	for i := 0; true; i++ {
		// fmt.Println("fdfd")
		if ok := vid.Read(&frame); !ok {
			fmt.Println("Frame Read completed")
			break
		}

		if i >= cfg.Limit {
			break
		}

		// if i < 40000 {
		// 	continue
		// }

		var laplacianScore = getLaplacianScore(&frame)

		fmt.Printf("\r Frame Cnt : %d  lscore %f", i, laplacianScore)

		// if maxLaplacianScore < laplacianScore && laplacianScore > 100 {
		// 	maxLaplacianScore = laplacianScore
		// 	lastMaxFrame = (frame.Clone())

		// }

		// if isInvalidImage(laplacianScore) {

		val := store(&frame, cfg)

		if val {
			fmt.Printf("Stored image %d  %.f\n ", ptr, getLaplacianScore(&frame))
		} else {
			fmt.Println("fail")
		}

		// maxLaplacianScore = math.SmallestNonzeroFloat64
		// lastMaxFrame = gocv.NewMat()
		// }
		// }

	}

	fmt.Println("Done...")

}
