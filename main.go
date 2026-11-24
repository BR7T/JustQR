package main

import (
	"SimpleQR/function"
	_ "embed"
	"log"
	"os"
	"fyne.io/systray"
)

//go:embed icon.ico
var iconData []byte

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--select" {
		RunSelect(os.Args[2])
		return
	}

	function.TurnDpiAware()

	f , err := function.InitLog()
	if err != nil{
		return
	}else{
		defer f.Close()
	}
	log.Println("app started")
	systray.Run(OnReady , onExit)
}


func OnReady() {
	systray.SetIcon(iconData)
	systray.SetTitle("SimpleQR")
	systray.SetTooltip("SimpleQR Run")

	mOpen := systray.AddMenuItem("Open", "Something")
	systray.AddSeparator()
	mExit := systray.AddMenuItem("Close", "Close SimpleQR")

	function.Shortcut()

	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				log.Println("Open Clicked")
			case <-mExit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	// limpeza (fechar arquivos, conexões etc.)
}
