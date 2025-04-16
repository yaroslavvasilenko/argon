package image

import "github.com/davidbyttow/govips/v2/vips"

func InitLibvips() {
	vips.Startup(nil)
}
