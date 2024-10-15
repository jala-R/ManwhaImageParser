package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"

	"gocv.io/x/gocv"
)

const configFilePath = "./config.json"

var ptr = 1

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
	Limit  int    `json: "limit"`
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
	var cfg = configParser()
	fmt.Println(cfg)
	if cfg.Stage == "test" {
		experiment(cfg)
	} else if cfg.Stage == "prod" {
		prod(cfg)
	}
}

func isInvalidImage(lap float64) bool {
	return lap < 15
}

func store(frame *gocv.Mat, cfg ConfigFile) bool {

	if frame == nil {
		return false
	}
	if frame.Empty() {
		return false
	}

	val := gocv.IMWrite(fmt.Sprintf("%s/%d.png", cfg.Output, ptr), *frame)
	ptr++

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

	maxLaplacianScore := math.SmallestNonzeroFloat64
	lastMaxFrame := gocv.NewMat()
	defer lastMaxFrame.Close()

	if cfg.Offset != 0 {
		vid.Grab(cfg.Offset)
	}

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

		if maxLaplacianScore < laplacianScore && laplacianScore > 25 {
			maxLaplacianScore = laplacianScore
			lastMaxFrame = (frame.Clone())

		}

		if isInvalidImage(laplacianScore) {

			val := store(&lastMaxFrame, cfg)

			if val {
				fmt.Printf("Stored image %d  %.f\n ", ptr, getLaplacianScore(&lastMaxFrame))
			}

			maxLaplacianScore = math.SmallestNonzeroFloat64
			lastMaxFrame = gocv.NewMat()
		}
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
