# JustQR

A system tray app that reads QR codes from any area of your screen with a hotkey.

No need to copy the image or use a website to decode it: the app runs in the background, you select the area, and it decodes the QR code.

## How it works

1. The app stays in the system tray.
2. Press `Win + Shift + Q`.
3. The screen is captured and dimmed. Drag to select the area with the QR code.
4. The selection is decoded and the result is logged.

Press `Esc` or right-click to cancel.

## Features

- Global hotkey
- Multi-monitor support (selection over the whole virtual screen)
- DPI aware
- Selection overlay with rounded corners and highlighted area
- Runs in the background from the system tray

## Requirements

- Windows
- [Go](https://go.dev/dl/) 1.21+

## Build

```
go build -ldflags "-H=windowsgui" -o JustQR.exe
```

The `-H=windowsgui` flag hides the console window.

## Project structure

- `main.go`: tray app entry point
- `select.go`: selection overlay (Win32)
- `function/`: screenshot, hotkey, logging and QR decoding

The same executable has two roles. Without arguments it runs the tray app. With `--select <image.png>` it opens the selection overlay, prints the selected rectangle as `x,y,w,h` and exits. Running the overlay in a separate process avoids conflicts between the tray and window message loops.

## Libraries

- [fyne.io/systray](https://github.com/fyne-io/systray)
- [golang.design/x/hotkey](https://github.com/golang-design/hotkey)
- [lxn/win](https://github.com/lxn/win)
- [gozxing](https://github.com/makiuchi-d/gozxing)
- [kbinani/screenshot](https://github.com/kbinani/screenshot)

## Status

Early development.
