package utils

import (
	"fmt"
)

type ByteBuffer struct {
	data []byte
	pos  int
}

func NewByteBuffer(data []byte) *ByteBuffer {
	buffer := &ByteBuffer{}
	if len(data) > 0 {
		buffer.data = append(buffer.data, data...)
	}
	buffer.pos = 0
	return buffer
}
func (this *ByteBuffer) Reset() {
	this.pos = 0
}
func (this *ByteBuffer) Skip(length int) bool {
	if (this.pos + length) > len(this.data) {
		return false
	}
	this.pos += length
	return true
}
func (this *ByteBuffer) ReadByte() (byte, bool) {
	if this.pos >= len(this.data) {
		return 0, false
	}
	res := this.data[this.pos]
	this.pos++
	return res, true
}
func (this *ByteBuffer) ReadShort() (int, bool) {
	if (this.pos + 2) > len(this.data) {
		return 0, false
	}
	res := (int)(this.data[this.pos]) + ((int)(this.data[this.pos + 1]) << 8)
	this.pos += 2
	return res, true
}
func (this *ByteBuffer) ReadBytes(length int) ([]byte, bool) {
	if this.pos >= len(this.data) {
		return []byte{}, false
	}
	res := make([]byte, length)
	copy(res, this.data[this.pos:(this.pos + length)])
	this.pos += length
	return res, true
}
func (this *ByteBuffer) ReadString(length int) (string, bool) {
	if (this.pos + length) > len(this.data) {
		return "", false
	}
	res := ""
	for i := 0; i < length; i++ {
		chr, _ := this.ReadByte()
		res += fmt.Sprintf("%c", chr)
	}
	return res, true
}

func (this *ByteBuffer) GetData() []byte {
	return this.data
}
func (this *ByteBuffer) AddByte(value byte) {
	this.data = append(this.data, value)
}
func (this *ByteBuffer) AddShort(value int) {
	data := []byte{(byte)(value & 0xFF), (byte)((value >> 8) & 0xFF)}
	this.data = append(this.data, data...)
}
func (this *ByteBuffer) AddBytes(data []byte) {
	this.data = append(this.data, data...)
}
func (this *ByteBuffer) AddString(str string) {
	if len(str) > 0 {
		data := []byte(str)
		this.data = append(this.data, data...)
	}
}
