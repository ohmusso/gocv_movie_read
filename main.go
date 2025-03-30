package main

import (
	"fmt"
	"readMovie/reader"
	"sync"
	"time"

	"gocv.io/x/gocv"
)

type Writer struct {
	videoWriter *gocv.VideoWriter
	ch          chan gocv.Mat
	frameNum    int
	wg          *sync.WaitGroup
}

func createWriter(path string, codec string, fps float64, width int, height int, frameNum int, wg *sync.WaitGroup) *Writer {
	videoWriter, err := gocv.VideoWriterFile(path, codec, fps, width, height, true)
	if err != nil {
		fmt.Println("write video file can't open")
		return nil
	}

	ch := make(chan gocv.Mat)

	return &Writer{videoWriter: videoWriter, ch: ch, frameNum: frameNum, wg: wg}
}

func (writer Writer) writeTask() {
	fmt.Println("start goroutine: writeTask")

	writer.wg.Add(1)

	index := 0
	timeSum := time.Duration(0)
	for mat := range writer.ch {
		index++

		s := time.Now()
		writer.videoWriter.Write(mat)
		e := time.Now()

		d := e.Sub(s)
		timeSum += d
	}
	avg := calcAve(timeSum.Milliseconds(), int64(index))
	fmt.Println("write time avg: ", avg, "ms")

	writer.wg.Done()
}

func (writer Writer) close() {
	writer.videoWriter.Close()
}

func calcAve(sum int64, num int64) float32 {
	sumF := float32(sum)
	numF := float32(num)
	avg := (sumF / numF)
	return avg
}

func main() {
	const channles = 3

	r := reader.NewMovieReader("./test.mp4", channles)
	defer r.Capture.Close()

	codec := r.GetCodec()
	fmt.Println("codec", codec)

	var wg sync.WaitGroup

	wrirtePathRed := "./test_red.mp4"
	writerRed := createWriter(wrirtePathRed, "avc1", r.Fps, r.Width, r.Height, r.FlameNum, &wg)
	if writerRed == nil {
		return
	}
	defer writerRed.close()

	wrirtePathGreen := "./test_green.mp4"
	writerGreen := createWriter(wrirtePathGreen, "avc1", r.Fps, r.Width, r.Height, r.FlameNum, &wg)
	if writerGreen == nil {
		return
	}
	defer writerGreen.close()

	wrirtePathBlue := "./test_blue.mp4"
	writerBlue := createWriter(wrirtePathBlue, "avc1", r.Fps, r.Width, r.Height, r.FlameNum, &wg)
	if writerBlue == nil {
		return
	}
	defer writerBlue.close()

	img := gocv.NewMat()
	defer img.Close()

	go writerRed.writeTask()
	go writerGreen.writeTask()
	go writerBlue.writeTask()

	start := time.Now()
	index := 0
	capReadTimeSum := time.Duration(0)
	newRGBFlameSum := time.Duration(0)

	frameCh := make(chan gocv.Mat)
	go func() {
		fmt.Println("start goroutine: frameCh")
		for frame := range frameCh {

			newRGBFlameStartTime := time.Now()

			rgb := reader.NewRGBFlame(&frame, channles)
			writerRed.ch <- rgb.Red
			writerGreen.ch <- rgb.Green
			writerBlue.ch <- rgb.Blue

			newRGBFlameEndTime := time.Now()
			newRGBFlameSum += newRGBFlameEndTime.Sub(newRGBFlameStartTime)
		}

		close(writerRed.ch)
		close(writerGreen.ch)
		close(writerBlue.ch)
	}()

	for {
		index++

		capReadStartTime := time.Now()
		isOk := r.Capture.Read(&img)

		if !isOk {
			fmt.Println("read end or error")
			break
		}

		if img.Empty() {
			fmt.Println("image empty")
			break
		}
		im := img.Clone()
		capReadEndTime := time.Now()

		frameCh <- im

		dCapRead := capReadEndTime.Sub(capReadStartTime)
		capReadTimeSum += dCapRead
	}
	fmt.Println("read frame done")
	close(frameCh)

	end := time.Now()

	avgCapRead := calcAve(int64(capReadTimeSum.Milliseconds()), int64(index))
	fmt.Println("cap read time avg: ", avgCapRead, "ms")

	avg := calcAve(int64(newRGBFlameSum.Milliseconds()), int64(index))
	fmt.Println("newRGBFlame time avg: ", avg, "ms")

	elapsed := end.Sub(start)
	fmt.Println("Total: ", elapsed.Milliseconds(), "ms")

	wg.Wait()
}
