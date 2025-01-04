package main

import (
	"readMovie/reader"
	"sync"
	"time"

	"gocv.io/x/gocv"
)

type Writer struct {
	videoWriter *gocv.VideoWriter
	ch          chan gocv.Mat
	wg          *sync.WaitGroup
}

func createWriter(path string, codec string, fps float64, width int, height int, wg *sync.WaitGroup) *Writer {
	videoWriter, err := gocv.VideoWriterFile(path, codec, fps, width, height, true)
	if err != nil {
		println("write video file can't open")
		return nil
	}

	ch := make(chan gocv.Mat)

	return &Writer{videoWriter: videoWriter, ch: ch, wg: wg}
}

func (writer Writer) taskWrite() {

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
	println("write time avg: ", avg)

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
	println("codec", codec)

	var wg sync.WaitGroup

	wrirtePathRed := "./test_red.mp4"
	writerRed := createWriter(wrirtePathRed, "avc1", r.Fps, r.Width, r.Height, &wg)
	if writerRed == nil {
		return
	}
	defer writerRed.close()

	wrirtePathGreen := "./test_green.mp4"
	writerGreen := createWriter(wrirtePathGreen, "avc1", r.Fps, r.Width, r.Height, &wg)
	if writerGreen == nil {
		return
	}
	defer writerGreen.close()

	wrirtePathBlue := "./test_blue.mp4"
	writerBlue := createWriter(wrirtePathBlue, "avc1", r.Fps, r.Width, r.Height, &wg)
	if writerBlue == nil {
		return
	}
	defer writerBlue.close()

	img := gocv.NewMat()
	defer img.Close()

	go writerRed.taskWrite()
	go writerGreen.taskWrite()
	go writerBlue.taskWrite()

	start := time.Now()
	index := 0
	readTimeSum := time.Duration(0)
	for {
		index++

		s := time.Now()
		isOk := r.Capture.Read(&img) // 10ms

		if !isOk {
			println("read end or error")
			break
		}

		if img.Empty() {
			println("image empty")
			break
		}

		rgb := reader.NewRGBFlame(&img, channles)
		writerRed.ch <- rgb.Red
		writerGreen.ch <- rgb.Green
		writerBlue.ch <- rgb.Blue

		e := time.Now()
		d := e.Sub(s)
		readTimeSum += d
	}

	end := time.Now()
	elapsed := end.Sub(start)
	println(elapsed.Milliseconds())

	avg := calcAve(int64(readTimeSum.Milliseconds()), int64(index))
	println("read time avg: ", avg)

	close(writerRed.ch)
	close(writerGreen.ch)
	close(writerBlue.ch)

	wg.Wait()
}
