package ui

import "image/color"

var (
	BlackColor  = color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF}
	WhiteColor  = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	GrayColor   = color.NRGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xFF}
	RedColor    = color.NRGBA{R: 0xB8, G: 0x40, B: 0x40, A: 0xFF}
	YellowColor = color.NRGBA{R: 0xB8, G: 0xB8, B: 0x40, A: 0xFF}
	GreenColor  = color.NRGBA{R: 0x40, G: 0xB8, B: 0x40, A: 0xFF}
	BlueColor   = color.NRGBA{R: 0x40, G: 0x40, B: 0xB8, A: 0xFF}
)

func Transparentize(c color.NRGBA, percent float32) color.NRGBA {
	if percent < 0 {
		percent = 0
	}
	if percent > 1 {
		percent = 1
	}
	a := float32(c.A)
	a *= percent
	c.A = uint8(a)
	return c
}

/************************************************/
/* snatched from Gio, yet another (p)oop moment */
/************************************************/

// MulAlpha applies the alpha to the color.
func MulAlpha(c color.NRGBA, alpha uint8) color.NRGBA {
	c.A = uint8(uint32(c.A) * uint32(alpha) / 0xFF)
	return c
}

// Disabled blends color towards the luminance and multiplies alpha.
// Blending towards luminance will desaturate the color.
// Multiplying alpha blends the color together more with the background.
func Disabled(c color.NRGBA) (d color.NRGBA) {
	const r = 80 // blend ratio
	lum := approxLuminance(c)
	d = mix(c, color.NRGBA{A: c.A, R: lum, G: lum, B: lum}, r)
	d = MulAlpha(d, 128+32)
	return
}

// approxLuminance is a fast approximate version of RGBA.Luminance.
func approxLuminance(c color.NRGBA) byte {
	const (
		r = 13933 // 0.2126 * 256 * 256
		g = 46871 // 0.7152 * 256 * 256
		b = 4732  // 0.0722 * 256 * 256
		t = r + g + b
	)
	return byte((r*int(c.R) + g*int(c.G) + b*int(c.B)) / t)
}

// mix mixes c1 and c2 weighted by (1 - a/256) and a/256 respectively.
func mix(c1, c2 color.NRGBA, a uint8) color.NRGBA {
	ai := int(a)
	return color.NRGBA{
		R: byte((int(c1.R)*ai + int(c2.R)*(256-ai)) / 256),
		G: byte((int(c1.G)*ai + int(c2.G)*(256-ai)) / 256),
		B: byte((int(c1.B)*ai + int(c2.B)*(256-ai)) / 256),
		A: byte((int(c1.A)*ai + int(c2.A)*(256-ai)) / 256),
	}
}
