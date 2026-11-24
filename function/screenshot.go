package function

import (
	"errors"
	"image"

	"github.com/kbinani/screenshot"
)
func Screenshot()(*image.RGBA , image.Rectangle ,error){
	n := screenshot.NumActiveDisplays()
	if n==0{
		return nil , image.Rectangle{} , errors.New("nenhum monitor")
	}	
	virtual := screenshot.GetDisplayBounds(0)
	for i := 1 ; i < n; i++{
		virtual = virtual.Union(screenshot.GetDisplayBounds(i))
	}

	img , err := screenshot.CaptureRect(virtual)
	if err != nil {
		return nil , virtual , err
	} 

	return img , virtual , nil
}
