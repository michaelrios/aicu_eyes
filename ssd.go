// What it does:
//
// This example shows, how to use pretrained SSD (Single Shot Detection) detection networks in gocv.
// Here, we detect human faces from the camera, but the setup is similar for any other kind of object detection.
//
// Download the (small, 5.1mb) Caffe model file from:
// https://raw.githubusercontent.com/opencv/opencv_3rdparty/dnn_samples_face_detector_20180205_fp16/res10_300x300_ssd_iter_140000_fp16.caffemodel
//
// Also, you will need the prototxt file:
// https://raw.githubusercontent.com/opencv/opencv/master/samples/dnn/face_detector/deploy.prototxt
//
// How to run:
//
// 		go run ./cmd/ssd-facedetect/main.go 0 [protofile] [modelfile]
//
// +build example

package main

import (
	"fmt"
	"github.com/EdlinOrg/prominentcolor"
	"github.com/disintegration/imaging"
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"runtime/trace"
	"strconv"
	"time"
	"gobot.io/x/gobot/platforms/raspi"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/gpio"
)

func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("How to run:\nssd-facedetect [camera ID] [protofile] [modelfile]")
		return
	}

	raspi := raspi.NewAdaptor()
	gpioreg.Register(gpio.PinIO())

	/////////////////////////////
	f, err := os.Create("trace.out")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = trace.Start(f)
	if err != nil {
		panic(err)
	}
	defer trace.Stop()
	/////////////////////////////

	// parse args
	deviceID, _ := strconv.Atoi(os.Args[1])
	proto := os.Args[2]
	model := os.Args[3]

	// open capture device
	webcam, err := gocv.VideoCaptureDevice(int(deviceID))
	if err != nil {
		fmt.Printf("Error opening video capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	window := gocv.NewWindow("SSD Face Detection")
	defer window.Close()

	img := gocv.NewMat()
	defer img.Close()

	// open DNN classifier
	net := gocv.ReadNetFromCaffe(proto, model)
	if net.Empty() {
		fmt.Printf("Error reading network model from : %v %v\n", proto, model)
		return
	}
	defer net.Close()

	green := color.RGBA{0, 255, 0, 0}
	fmt.Printf("Start reading camera device: %v\n", deviceID)

	var c, r int
	var W, H, left, right, top, bottom, confidence float32
	var errImg error

	var detections, detBlob, blob gocv.Mat
	var originalImg image.Image
	var centroidColor color.RGBA

	prominentcolor.MaskBlack = prominentcolor.MaskWhite
	prominentcolor.MaskGreen = prominentcolor.MaskWhite

	for i := 0; i < 500; i++ {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Error cannot read device %d\n", deviceID)
			return
		}

		if img.Empty() {
			continue
		}

		W = float32(img.Cols())
		H = float32(img.Rows())

		// convert image Mat to 96x128 blob that the detector can analyze
		blob = gocv.BlobFromImage(img, 1.0, image.Pt(128, 96), gocv.NewScalar(104.0, 177.0, 123.0, 0), false, false)
		defer blob.Close()

		// feed the blob into the classifier
		net.SetInput(blob, "data")

		// run a forward pass through the network
		detBlob = net.Forward("detection_out")
		defer detBlob.Close()

		// extract the detections.
		// for each object detected, there will be 7 float features:
		// objid, classid, confidence, left, top, right, bottom.
		detections = gocv.GetBlobChannel(detBlob, 0, 0)
		defer detections.Close()

		originalImg, errImg = img.ToImage()
		if errImg != nil {
			log.Fatalf("Failed converting Mat to Image: %s", errImg.Error())
		}

		for r = 0; r < detections.Rows(); r++ {
			// you would want the classid for general object detection,
			// but we do not need it here.
			// classid := detections.GetFloatAt(r, 1)

			confidence = detections.GetFloatAt(r, 2)
			if confidence < 0.4 {
				continue
			}

			left = detections.GetFloatAt(r, 3) * W
			top = detections.GetFloatAt(r, 4) * H
			right = detections.GetFloatAt(r, 5) * W
			bottom = detections.GetFloatAt(r, 6) * H

			// scale to video size:
			left = min(max(0, left-20), W-1)
			right = min(max(0, right+20), W-1)
			bottom = min(max(0, bottom+100), H-1)
			top = min(max(0, bottom+200), H-1)

			// draw it
			//fmt.Printf("left %d bottom %d right %d bottom %d \n", int(left), int(top), int(right), int(bottom))
			rect := image.Rect(int(left), int(top), int(right), int(bottom))
			gocv.Rectangle(&img, rect, green, 1)

			croppedImg := imaging.Crop(originalImg, rect)

			centroids, err := prominentcolor.KmeansWithArgs(prominentcolor.ArgumentNoCropping|prominentcolor.ArgumentAverageMean, croppedImg)
			if err != nil {
				fmt.Printf("Error get colors: %s \n", err.Error())
			}

			for c = 0; c < len(centroids); c++ {
				centroidColor = color.RGBA{uint8(centroids[c].Color.R), uint8(centroids[c].Color.G), uint8(centroids[c].Color.B), 0}

				hsl := convertRGBToHSL(centroidColor)

				gocv.Rectangle(&img, rect, centroidColor, 30-((c+1)*10))
				gocv.PutText(&img, hsl.Classify(), image.Pt(10, 20*(c+1)), gocv.FontHersheyPlain, 2, green, 2)

				fmt.Print(convertRGBToHex(centroidColor), " ")
			}
		}
		fmt.Println()

		window.IMShow(img)
		if window.WaitKey(1) >= 0 {
			break
		}
	}
}

type Customer struct {
	Id            string
	PerceivedTeam string
	ShirtColors   []ShirtColor
	StartTime     time.Time
	EndTime       time.Time
}

type ShirtColor struct {
	Color  string  `json:"color"`
	Rating float32 `json:"rating"`
}

type HSL struct {
	H float32
	S float32
	L float32
}

func (hsl HSL) Classify() string {
	if hsl.L < 0.2 {
		return "black"
	}
	if hsl.L > 0.8 {
		return "white"
	}

	if hsl.S < 0.2 {
		return "gray"
	}

	if hsl.H < 30 {
		return "red"
	}
	if hsl.H < 90 {
		return "yellow"
	}
	if hsl.H < 170 {
		return "green"
	}
	if hsl.H < 270 {
		return "blue"
	}
	if hsl.H < 330 {
		return "magenta"
	}
	return "red"
}

func convertRGBToHSL(rgb color.RGBA) (hsl HSL) {
	r := float32(rgb.R) / 255
	g := float32(rgb.G) / 255
	b := float32(rgb.B) / 255

	max := max(r, max(g, b))
	min := min(r, min(g, b))

	delta := max - min

	var hue float32
	if max == r {
		hue = (g - b) / delta
	} else if max == g {
		hue = 2 + ((b - r) / delta)
	} else {
		hue = 4 + ((r - g) / delta)
	}

	hsl.H = hue * 60

	hsl.L = (max + min) / 2

	hsl.S = float32(delta) / (1 - float32(math.Abs(float64((2*hsl.L)-1))))

	return hsl
}

func convertRGBToHex(rgb color.RGBA) string {
	return fmt.Sprintf("%x%x%x", rgb.R, rgb.G, rgb.B)
}
