//go:build desktop

package gui

import (
	"bytes"
	"encoding/binary"
	_ "embed"
)

//go:embed assets/tray_icon.png
var trayIconPNG []byte

// pngToICO wraps PNG bytes into a minimal ICO container (PNG-in-ICO format,
// supported on Windows Vista+). This avoids needing an external .ico file.
func pngToICO(png []byte) ([]byte, error) {
	var buf bytes.Buffer

	// ICONDIR header (6 bytes)
	buf.Write([]byte{0x00, 0x00}) // reserved
	buf.Write([]byte{0x01, 0x00}) // type = 1 (ICO)
	buf.Write([]byte{0x01, 0x00}) // count = 1 image

	// ICONDIRENTRY (16 bytes)
	buf.WriteByte(0) // width (0 = 256)
	buf.WriteByte(0) // height (0 = 256)
	buf.WriteByte(0) // color count (0 = no palette)
	buf.WriteByte(0) // reserved
	binary.Write(&buf, binary.LittleEndian, uint16(0))    // color planes
	binary.Write(&buf, binary.LittleEndian, uint16(32))   // bits per pixel
	binary.Write(&buf, binary.LittleEndian, uint32(len(png))) // image size
	binary.Write(&buf, binary.LittleEndian, uint32(22))   // offset (6 + 16 = 22)

	// PNG image data
	buf.Write(png)

	return buf.Bytes(), nil
}

func trayIconBytes() []byte {
	ico, err := pngToICO(trayIconPNG)
	if err != nil {
		return trayIconPNG
	}
	return ico
}
