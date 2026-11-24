package function

import (
	"log"
	"os"
	"path/filepath"
)

func InitLog() (*os.File , error){
	base , err := os.UserConfigDir()
	if err != nil{
		return nil , err
	}
	dir := filepath.Join(base , "SimpleQR")
	if err := os.MkdirAll(dir, 0o755); err != nil{
		return nil , err
	}
	f , err := os.OpenFile(
		filepath.Join(dir, "app.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0o644,
	)
	if err != nil {
		return  nil , err
	}
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	return f , nil
}
