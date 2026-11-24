package function

import (
	"image"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

func DecodeQR(img image.Image)(string , error){
	bmp , err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "" , err
	}
	res , err := qrcode.NewQRCodeReader().Decode(bmp , nil)
	if err != nil {
		return "" , err
	}
	return res.GetText() , nil
}
