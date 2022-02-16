package main

import (
	"./utils"
	"fmt"
	"os"
	"time"
)

type Color struct {
	R byte
	G byte
	B byte
}
type DataBlock []byte

type ImageData struct {
	left     int
	top      int
	width    int
	height   int
	lctFlag  byte
	ilFlag   byte
	sortFlag byte
	sizeLCT  byte
	clrs     []Color

	lzwMinCodeSize byte
	blockDatas     []DataBlock
}

type GraphicControl struct {
	byteSize byte // 固定：0x04
	delayTime int
	trClrIdx  byte

	disposalMethod byte
	userInputFlag  byte
	transFlag      byte
}

type Frame struct {
	grpCtrl *GraphicControl
	imgData *ImageData
}

type GIF struct {
	verion   string
	width    int
	height   int
	gctFlag  byte
	clrRes   byte
	sortFlag byte
	sizeGCT  byte
	bgClrIdx byte
	pixAR    byte
	glbClrs  []Color

	frames   []*Frame
}

func main() {
	filename := "demo.gif"
	filedata := utils.ReadFile(filename)
	gifData := utils.NewByteBuffer(filedata)
	gif := parseFile(gifData)
	if gif == nil {
		return
	}
	//todo
	println("Size: ", gif.width, " x ", gif.height, " frames: ", len(gif.frames))
	saveFrames(gif)
}

func saveFrames(gif *GIF) {
	now := time.Now()
	folder := "./" + now.Format("06-01-02 15-04-05")
	os.MkdirAll(folder, 0777)

	for i := 0; i < len(gif.frames); i++ {
		filename := fmt.Sprintf("%s/%d.gif", folder, i) 
		saveFrame(gif, gif.frames[i], filename)
	}
}

func saveFrame(gif *GIF, frame *Frame, filename string) {
	nullData := []byte{}
	newBuffer := utils.NewByteBuffer(nullData)
	newBuffer.AddString("GIF")
	newBuffer.AddString(gif.verion)
	// Logical Screen Descriptor
	newBuffer.AddShort(gif.width)
	newBuffer.AddShort(gif.height)
	packF := (gif.gctFlag << 7) + (gif.clrRes << 4) + (gif.sortFlag << 3) + gif.sizeGCT
	newBuffer.AddByte(packF)
	newBuffer.AddByte(gif.bgClrIdx)
	newBuffer.AddByte(gif.pixAR)
	// Global Color Table
	for i := 0; i < len(gif.glbClrs); i++ {
		color := gif.glbClrs[i]
		newBuffer.AddByte(color.R)
		newBuffer.AddByte(color.G)
		newBuffer.AddByte(color.B)
	}

	// Graphic Cotroll Extension
	if frame.grpCtrl != nil {
		newBuffer.AddByte(0x21)
		newBuffer.AddByte(0xF9)
		newBuffer.AddByte(frame.grpCtrl.byteSize)
		packF = (frame.grpCtrl.disposalMethod << 2) + (frame.grpCtrl.userInputFlag << 1) + frame.grpCtrl.transFlag
		newBuffer.AddByte(packF)
		newBuffer.AddShort(frame.grpCtrl.delayTime)
		newBuffer.AddByte(frame.grpCtrl.trClrIdx)
		newBuffer.AddByte(0x00)
	}

	// Image Data
	newBuffer.AddByte(0x2C)
	newBuffer.AddShort(frame.imgData.left)
	newBuffer.AddShort(frame.imgData.top)
	newBuffer.AddShort(frame.imgData.width)
	newBuffer.AddShort(frame.imgData.height)
	packF = (frame.imgData.lctFlag << 7) + (frame.imgData.ilFlag << 6) + (frame.imgData.sortFlag << 5) + frame.imgData.sizeLCT
	newBuffer.AddByte(packF)
	for i := 0; i < len(frame.imgData.clrs); i++ {
		color := frame.imgData.clrs[i]
		newBuffer.AddByte(color.R)
		newBuffer.AddByte(color.G)
		newBuffer.AddByte(color.B)
	}
	newBuffer.AddByte(frame.imgData.lzwMinCodeSize)
	for i := 0; i < len(frame.imgData.blockDatas); i++ {
		blockData := frame.imgData.blockDatas[i]
		newBuffer.AddByte((byte)(len(blockData)))
		newBuffer.AddBytes(blockData)
	}
	newBuffer.AddByte(0x3B)

	utils.WriteFile(filename, newBuffer.GetData())
}

func parseFile(data *utils.ByteBuffer) *GIF {
	// Header
	headStr, _ := data.ReadString(3)
	if headStr != "GIF" {
		println("Not a GIF image")
		return nil
	}

	gif := &GIF{}
	gif.verion, _ = data.ReadString(3)

	// Logical Screen Descriptor
	gif.width, _  = data.ReadShort()
	gif.height, _ = data.ReadShort()

	packF, _ := data.ReadByte()
	gif.gctFlag  = (packF >> 7)
	gif.clrRes   = (packF >> 4) & 0x7
	gif.sortFlag = (packF >> 3) & 0x1
	gif.sizeGCT  = (packF) & 0x7

	gif.bgClrIdx, _ = data.ReadByte()
	gif.pixAR, _    = data.ReadByte()
	// Global Color Table
	if gif.gctFlag > 0 {
		clrCount := (0x1 << (gif.sizeGCT + 1))
		for i := 0; i < clrCount; i ++ {
			r, _ := data.ReadByte()
			g, _ := data.ReadByte()
			b, _ := data.ReadByte()
			gif.glbClrs = append(gif.glbClrs, Color{r, g, b})
		}
	}

	for {
		flag, _ := data.ReadByte()
		if flag == 0x3B { // End Of File
			break
		}
		if flag == 0x21 {
			secFlag, _ := data.ReadByte()
			switch (secFlag) {
				case 0x01: { // Plain Text Extension
					parsePlainText(data)
					break
				}
				case 0xF9: { // Graphic Control Extension
					parseGraphicControl(data, gif)
					break
				}
				case 0xFE: { // Comment Extension
					parseComment(data)
					break
				}
				case 0xFF: { // Application Extension
					parseApplication(data)
					break
				}
			}
		} else if flag == 0x2C { // Image Descriptor
			parseImage(data, gif)
		}
	}
	return gif
}

func parseGraphicControl(data *utils.ByteBuffer, gif *GIF) {
	grpCtrl := &GraphicControl{}
	grpCtrl.byteSize, _ = data.ReadByte()
	packField, _ := data.ReadByte()
	grpCtrl.delayTime, _ = data.ReadShort()
	grpCtrl.trClrIdx, _ = data.ReadByte()
	data.ReadByte() // Block Terminator

	grpCtrl.disposalMethod = (packField >> 2) & 0x7
	grpCtrl.userInputFlag = (packField >> 1) & 0x1
	grpCtrl.transFlag = (packField) & 0x1

	frame := &Frame{
		grpCtrl: grpCtrl,
	}
	gif.frames = append(gif.frames, frame)
}
func parseImage(data *utils.ByteBuffer, gif *GIF) {
	imgData := &ImageData{}
	imgData.left, _   = data.ReadShort()
	imgData.top, _    = data.ReadShort()
	imgData.width, _  = data.ReadShort()
	imgData.height, _ = data.ReadShort()
	packF, _  := data.ReadByte()
	imgData.lctFlag  = (packF >> 7)
	imgData.ilFlag   = (packF >> 6) & 0x1
	imgData.sortFlag = (packF >> 5) & 0x1
	imgData.sizeLCT  = (packF) & 0x7
	if imgData.lctFlag > 0 { // Local Color Table
		clrCount := (0x1 << (imgData.sizeLCT + 1))
		for i := 0; i < clrCount; i ++ {
			r, _ := data.ReadByte()
			g, _ := data.ReadByte()
			b, _ := data.ReadByte()
			imgData.clrs = append(imgData.clrs, Color{r, g, b})
		}
	}
	// Image Data
	imgData.lzwMinCodeSize, _ = data.ReadByte()
	for {
		blockSize, _ := data.ReadByte()
		if blockSize == 0 {
			break
		}
		blockData, _ := data.ReadBytes((int)(blockSize))
		imgData.blockDatas = append(imgData.blockDatas, blockData)
	}
	var frame *Frame = nil
	if len(gif.frames) == 0 {
		frame = &Frame{}
		gif.frames = append(gif.frames, frame)
	} else {
		frame = gif.frames[len(gif.frames) - 1]
		if frame.imgData != nil {
			frame = &Frame{}
			gif.frames = append(gif.frames, frame)
		}
	}
	frame.imgData = imgData
}
func parseApplication(data *utils.ByteBuffer) {
	data.Skip(1)
	appId, _ := data.ReadString(8)
	appAuthCode, _ := data.ReadString(3)
	println("Application: ", appId, " Auth Code: ", appAuthCode)

	blockDatas := []DataBlock{}
	for {
		blockSize, _ := data.ReadByte()
		if blockSize == 0 {
			break
		}
		blockData, _ := data.ReadBytes((int)(blockSize))
		blockDatas = append(blockDatas, blockData)
	}
	println("Application Extension: ", len(blockDatas), " blocks")
}
func parseComment(data *utils.ByteBuffer) {
	for {
		blockSize, _ := data.ReadByte()
		if blockSize == 0 {
			break
		}
		data.Skip((int)(blockSize))
	}
	println("Comment Extension")
	// blockDatas := []DataBlock{}
	// for {
	// 	blockSize, _ := data.ReadByte()
	// 	if blockSize == 0 {
	// 		break
	// 	}
	// 	blockData, _ := data.ReadBytes((int)(blockSize))
	// 	blockDatas = append(blockDatas, blockData)
	// }
	// println("Comment Extension: ", len(blockDatas), " blocks")
}
func parsePlainText(data *utils.ByteBuffer) {
	data.Skip(1)
	left, _    := data.ReadShort()
	top, _     := data.ReadShort()
	width, _   := data.ReadShort()
	height, _  := data.ReadShort()
	cWidth, _  := data.ReadByte()
	cHeight, _ := data.ReadByte()
	fgClrIdx, _ := data.ReadByte()
	bgClrIdx, _ := data.ReadByte()

	println("Plain Text Extension: ", left, top, width, height, cWidth, cHeight, fgClrIdx, bgClrIdx)

	blockDatas := []DataBlock{}
	for {
		blockSize, _ := data.ReadByte()
		if blockSize == 0 {
			break
		}
		blockData, _ := data.ReadBytes((int)(blockSize))
		blockDatas = append(blockDatas, blockData)
	}
	println("Plain Text Extension: ", len(blockDatas), " blocks")
}
