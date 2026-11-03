package mizan

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

// IconStroke, bütün ikonların ekrandaki çizgi kalınlığıdır (DIP). Kalınlık
// viewBox'tan ve boyuttan hesaplanır: her boyutta ekranda aynı ağırlık.
const IconStroke = 1.6

// Svg, Mizan'ın çizgi ikonudur: dolgusuz, metin renginde kontur, yuvarlak
// uçlar. viewBox 0 0 box box'tır; size ekrandaki genişliktir.
func Svg(box, size float32, paths ...string) *ui.SVG {
	src := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %g %g" width="%g" height="%g" fill="none" stroke="currentColor" stroke-width="%g" stroke-linecap="round" stroke-linejoin="round">`,
		box, box, size, size, IconStroke*box/size)
	for _, d := range paths {
		src += `<path d="` + d + `"/>`
	}
	s, err := ui.ParseSVG([]byte(src + `</svg>`))
	if err != nil {
		panic(err)
	}
	return s
}

var (
	iconChevrons = Svg(16, 14, "M5 6.5L8 3.5l3 3M5 9.5l3 3 3-3")
	iconCheck    = Svg(16, 14, "M3.5 8.5l3 3 6-7")
	iconClear    = Svg(16, 12, "M4.5 4.5l7 7M11.5 4.5l-7 7")
	iconEye      = Svg(16, 15, "M1.5 8C2.4 5.1 5.1 3 8 3s5.6 2.1 6.5 5c-.9 2.9-3.6 5-6.5 5S2.4 10.9 1.5 8z")
	iconEyeOff   = Svg(16, 15, "M2 2l12 12M6.6 6.6a2 2 0 0 0 2.8 2.8M4.2 4.4C2.8 5.3 1.9 6.6 1.5 8c.9 2.9 3.6 5 6.5 5 1.3 0 2.5-.4 3.6-1.1M7 3.1c.3 0 .7-.1 1-.1 2.9 0 5.6 2.1 6.5 5-.3 1-.8 1.9-1.5 2.6")
)
