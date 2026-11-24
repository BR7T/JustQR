package function

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"golang.design/x/hotkey"
)

var selected atomic.Bool

func Shortcut() {
	hk := hotkey.New([]hotkey.Modifier{hotkey.ModWin , hotkey.ModShift}, hotkey.KeyQ)
	if err := hk.Register(); err != nil{
		return
	}
	go func(){
		for range hk.Keydown(){
			if !selected.CompareAndSwap(false,true){
				continue
			}
			
			func(){
				defer selected.Store(false)
				img , _ , err := Screenshot()
				if err != nil {
					log.Println(err)
				}

				ret , ok := selectImg(img)
				if !ok {
					log.Println("selecção cancelada")
					return
				}

				txt , err := DecodeQR(img.SubImage(ret))
				if err != nil {
					log.Println("QR não encontrado:",err)
					return
				}
				log.Println("QR:" , txt)
				if !ok {
					log.Println("imagem sem SubImage")
					return
				}

				log.Println("retângulo:" , ret)
				 		

			}()
		}
	}()
}

func selectImg(img image.Image)(image.Rectangle , bool){
	tmp := filepath.Join(os.TempDir() , "cp.png")
	f, err := os.Create(tmp)
	if err != nil {
		log.Println(err)
		return image.Rectangle{} , false
	}

	err = png.Encode(f, img)
	f.Close()
	if err != nil {
		log.Println(err)
		return image.Rectangle{}, false
	}
	defer os.Remove(tmp)

	exe , err := os.Executable()
	if err != nil {
		log.Println(err)
		return image.Rectangle{} , false
	}

	out , err := exec.Command(exe , "--select" , tmp).Output()
	if err != nil || len(out) == 0 {
		return image.Rectangle{} , false
	}

	var x,y,w,h int
	if _, err := fmt.Sscanf(string(out) , "%d,%d,%d,%d" ,&x,&y,&w,&h);err != nil{
		log.Println(err)
		return image.Rectangle{} , false
	}
	return image.Rect(x,y,x+w,y+h) , true

}
