package main

import (
	"fmt"
	"math"
	"os"

	"gocv.io/x/gocv"
)

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

func main() {
	if os.Args[1] == "test" {
		experiment()
	} else if os.Args[1] == "prod" {
		prod()
	}
}

func isInvalidImage(lap float64) bool {
	return lap < 25
}

func store(frame *gocv.Mat) bool {

	if frame == nil {
		return false
	}
	if frame.Empty() {
		return false
	}

	val := gocv.IMWrite(fmt.Sprintf("%s/%d.png", os.Args[3], ptr), *frame)
	ptr++

	return val
}

func prod() {

	fmt.Println("Running")
	vid, err := gocv.VideoCaptureFile(os.Args[2])

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
	var lastMaxFrame gocv.Mat

	for i := 0; true; i++ {
		// fmt.Println("fdfd")
		if ok := vid.Read(&frame); !ok {
			fmt.Println("Frame Read completed")
			break
		}

		// if i < 40000 {
		// 	continue
		// }

		var laplacianScore = getLaplacianScore(&frame)

		fmt.Printf("\r Frame Cnt : %d  lscore %f", i, laplacianScore)

		if maxLaplacianScore < laplacianScore && laplacianScore > 50 {
			maxLaplacianScore = laplacianScore
			lastMaxFrame = (frame.Clone())

		}

		if isInvalidImage(laplacianScore) {

			val := store(&lastMaxFrame)

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

func experiment() {

	fmt.Println("Running")
	vid, err := gocv.VideoCaptureFile(os.Args[2])

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

	for i := 0; true; i++ {
		// fmt.Println("fdfd")
		if ok := vid.Read(&frame); !ok {
			fmt.Println("Frame Read completed")
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

		val := store(&frame)

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
