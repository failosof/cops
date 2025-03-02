package util

import (
	"image"
)

func Rect(origin, size image.Point) image.Rectangle {
	return image.Rectangle{
		Min: origin,
		Max: origin.Add(size),
	}
}
