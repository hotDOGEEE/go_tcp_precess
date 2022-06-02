package convert

import (
	"encoding/binary"

	lz4c "github.com/pierrec/lz4"
)

// 对应到协议的字段
var MapLit = map[string]map[string]int{
	"Handshake": {"1": 1, "2": 4, "3": 1},
}

func SliceRever(slice []byte) []byte {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
	return slice
}

func Uint16ToBytes(u uint16) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, u)
	return buf
}

func Uint32ToBytes(u uint32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, u)
	return buf
}

func BytesToUint32(b []byte) uint32 {
	return binary.BigEndian.Uint32(b)
}

func BytesToUint16(b []byte) uint16 {
	return binary.BigEndian.Uint16(b)
}

// 包含一个压缩方法，看项目是否使用
func UnzipLz4c(lens_data int, data []byte) ([]byte, error) {
	oriData := make([]byte, lens_data)
	size, err := lz4c.UncompressBlock(data, oriData, 0)
	if err != nil {
		return nil, err
	}
	return oriData[:size], nil
}
