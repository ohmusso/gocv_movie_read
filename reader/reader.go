package reader

import (
	"fmt"

	"gocv.io/x/gocv"
)

type MovieReader struct {
	Capture  *gocv.VideoCapture
	Channels int
	FlameNum int
	Fps      float64
	Height   int
	Width    int
}

func NewMovieReader(imgPath string, channels int) *MovieReader {
	p := new(MovieReader)

	fmt.Printf("gocv version: %s\n", gocv.Version())
	fmt.Printf("opencv lib version: %s\n", gocv.OpenCVVersion())
	capture, error := gocv.OpenVideoCapture(imgPath)

	if error != nil {
		fmt.Println(error)
		return nil
	}

	p.Capture = capture
	p.Channels = channels
	p.FlameNum = int(capture.Get(gocv.VideoCaptureFrameCount))
	p.Height = int(capture.Get(gocv.VideoCaptureFrameHeight))
	p.Width = int(capture.Get(gocv.VideoCaptureFrameWidth))
	p.Fps = capture.Get(gocv.VideoCaptureFPS)

	fmt.Println("VideoCaptureFPS", capture.Get(gocv.VideoCaptureFPS))
	fmt.Println("VideoCaptureFrameCount", p.FlameNum)
	fmt.Println("VideoCaptureFrameHeight", p.Height)
	fmt.Println("VideoCaptureFrameWidth", p.Width)

	return p
}

func (imgReader *MovieReader) closeImgReader() {
	if imgReader.Capture == nil {
		return
	}

	imgReader.Capture.Close()
}

func (imgReader *MovieReader) GetCodec() string {
	if imgReader.Capture == nil {
		return ""
	}

	return imgReader.Capture.CodecString()
}

type RGBFlame struct {
	Red   gocv.Mat
	Blue  gocv.Mat
	Green gocv.Mat
}

func NewRGBFlame(img *gocv.Mat, chNum int) *RGBFlame {
	rgbFlame := new(RGBFlame)
	rowNum := img.Rows()
	colNum := img.Cols()
	rgbFlame.Red = gocv.NewMatWithSize(img.Rows(), img.Cols(), gocv.MatTypeCV8UC3)
	rgbFlame.Green = gocv.NewMatWithSize(img.Rows(), img.Cols(), gocv.MatTypeCV8UC3)
	rgbFlame.Blue = gocv.NewMatWithSize(img.Rows(), img.Cols(), gocv.MatTypeCV8UC3)

	p, _ := img.DataPtrUint8()
	pBlue, _ := rgbFlame.Blue.DataPtrUint8()
	pGreen, _ := rgbFlame.Green.DataPtrUint8()
	pRed, _ := rgbFlame.Red.DataPtrUint8()
	total := 0
	for i := 0; i < rowNum; i++ {
		imgRowIndex := i * img.Cols() * chNum
		for j := 0; j < colNum; j++ {
			imgColIndex := j * chNum
			imgIndex := imgRowIndex + imgColIndex
			// BGR
			pBlue[imgIndex+0] = p[imgIndex+0]
			pGreen[imgIndex+1] = p[imgIndex+1]
			pRed[imgIndex+2] = p[imgIndex+2]
			total++
		}
	}

	return rgbFlame
}
