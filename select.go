package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"runtime"
	"syscall"
	"unsafe"

	"SimpleQR/function"

	"github.com/lxn/win"
)

const (
	dimDiv = 3  // divisor do brilho do fundo (2 = 50%)
	corner = 8 // diâmetro dos cantos arredondados
)

var (
	bmp, bmpDim, backBmp win.HBITMAP
	imgW, imgH           int32

	dragging                   bool
	startX, startY, curX, curY int32

	selected bool
	result   image.Rectangle
)

var (
	gdi32              = syscall.NewLazyDLL("gdi32.dll")
	procCreatePen      = gdi32.NewProc("CreatePen")
	procRoundRect      = gdi32.NewProc("RoundRect")
	procCreateRoundRgn = gdi32.NewProc("CreateRoundRectRgn")
	procSelectClipRgn  = gdi32.NewProc("SelectClipRgn")
)

func createPen(style, width int32, color uint32) win.HPEN {
	r, _, _ := procCreatePen.Call(uintptr(style), uintptr(width), uintptr(color))
	return win.HPEN(r)
}

func roundRect(hdc win.HDC, l, t, r, b, d int32) {
	procRoundRect.Call(uintptr(hdc), uintptr(l), uintptr(t), uintptr(r), uintptr(b), uintptr(d), uintptr(d))
}

func createRoundRgn(l, t, r, b, d int32) win.HRGN {
	rgn, _, _ := procCreateRoundRgn.Call(uintptr(l), uintptr(t), uintptr(r), uintptr(b), uintptr(d), uintptr(d))
	return win.HRGN(rgn)
}

func selectClipRgn(hdc win.HDC, rgn win.HRGN) {
	procSelectClipRgn.Call(uintptr(hdc), uintptr(rgn))
}

func loadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

func toRGBA(img image.Image) *image.RGBA {
	if rgba, ok := img.(*image.RGBA); ok {
		return rgba
	}
	b := img.Bounds()
	rgba := image.NewRGBA(b)
	draw.Draw(rgba, b, img, b.Min, draw.Src)
	return rgba
}

func createBitmap(img *image.RGBA, dim bool) win.HBITMAP {
	b := img.Bounds()
	w, h := int32(b.Dx()), int32(b.Dy())
	pix := make([]byte, int(w)*int(h)*4)
	i := 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := img.RGBAAt(x, y)
			r, g, bl := c.R, c.G, c.B
			if dim {
				r, g, bl = r/dimDiv, g/dimDiv, bl/dimDiv
			}
			pix[i], pix[i+1], pix[i+2], pix[i+3] = bl, g, r, 255
			i += 4
		}
	}
	return win.CreateBitmap(w, h, 1, 32, unsafe.Pointer(&pix[0]))
}

func virtualScreen() (x, y, w, h int32) {
	x = win.GetSystemMetrics(win.SM_XVIRTUALSCREEN)
	y = win.GetSystemMetrics(win.SM_YVIRTUALSCREEN)
	w = win.GetSystemMetrics(win.SM_CXVIRTUALSCREEN)
	h = win.GetSystemMetrics(win.SM_CYVIRTUALSCREEN)
	return
}

func normalize(x1, y1, x2, y2 int32) (l, t, r, b int32) {
	l, r = x1, x2
	if l > r {
		l, r = r, l
	}
	t, b = y1, y2
	if t > b {
		t, b = b, t
	}
	return
}

func wndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {

	case win.WM_ERASEBKGND:
		return 1

	case win.WM_PAINT:
		var ps win.PAINTSTRUCT
		hdc := win.BeginPaint(hwnd, &ps)

		if backBmp == 0 {
			backBmp = win.CreateCompatibleBitmap(hdc, imgW, imgH)
		}
		buf := win.CreateCompatibleDC(hdc)
		oldBuf := win.SelectObject(buf, win.HGDIOBJ(backBmp))
		mem := win.CreateCompatibleDC(hdc)

		// 1. fundo escurecido (em buf)
		old := win.SelectObject(mem, win.HGDIOBJ(bmpDim))
		win.BitBlt(buf, 0, 0, imgW, imgH, mem, 0, 0, win.SRCCOPY)

		if dragging {
			l, t, r, b := normalize(startX, startY, curX, curY)

			// 2. seleção clara, recortada com cantos arredondados
			rgn := createRoundRgn(l, t, r+1, b+1, corner)
			selectClipRgn(buf, rgn)
			win.SelectObject(mem, win.HGDIOBJ(bmp))
			win.BitBlt(buf, l, t, r-l, b-t, mem, l, t, win.SRCCOPY)
			selectClipRgn(buf, 0)
			win.DeleteObject(win.HGDIOBJ(rgn))

			// 3. borda arredondada
			pen := createPen(win.PS_SOLID, 2, 0x00FFFFFF) // BGR
			oldPen := win.SelectObject(buf, win.HGDIOBJ(pen))
			oldBrush := win.SelectObject(buf, win.GetStockObject(win.NULL_BRUSH))
			roundRect(buf, l, t, r, b, corner)
			win.SelectObject(buf, oldBrush)
			win.SelectObject(buf, oldPen)
			win.DeleteObject(win.HGDIOBJ(pen))
		}

		// única cópia para a tela
		win.BitBlt(hdc, 0, 0, imgW, imgH, buf, 0, 0, win.SRCCOPY)

		win.SelectObject(mem, old)
		win.DeleteDC(mem)
		win.SelectObject(buf, oldBuf)
		win.DeleteDC(buf)
		win.EndPaint(hwnd, &ps)
		return 0

	case win.WM_LBUTTONDOWN:
		startX, startY = win.GET_X_LPARAM(lParam), win.GET_Y_LPARAM(lParam)
		curX, curY = startX, startY
		dragging = true
		win.SetCapture(hwnd)
		return 0

	case win.WM_MOUSEMOVE:
		if dragging {
			curX, curY = win.GET_X_LPARAM(lParam), win.GET_Y_LPARAM(lParam)
			win.InvalidateRect(hwnd, nil, false)
		}
		return 0

	case win.WM_LBUTTONUP:
		if dragging {
			dragging = false
			win.ReleaseCapture()
			curX, curY = win.GET_X_LPARAM(lParam), win.GET_Y_LPARAM(lParam)
			l, t, r, b := normalize(startX, startY, curX, curY)
			if r-l >= 5 && b-t >= 5 {
				result = image.Rect(int(l), int(t), int(r), int(b))
				selected = true
			}
			win.DestroyWindow(hwnd)
		}
		return 0

	case win.WM_RBUTTONDOWN:
		win.DestroyWindow(hwnd)
		return 0

	case win.WM_KEYDOWN:
		if wParam == win.VK_ESCAPE {
			win.DestroyWindow(hwnd)
		}
		return 0

	case win.WM_DESTROY:
		win.DeleteObject(win.HGDIOBJ(bmp))
		win.DeleteObject(win.HGDIOBJ(bmpDim))
		win.DeleteObject(win.HGDIOBJ(backBmp))
		win.PostQuitMessage(0)
		return 0
	}
	return win.DefWindowProc(hwnd, msg, wParam, lParam)
}

func registerClass(name *uint16) {
	hInst := win.GetModuleHandle(nil)
	wc := win.WNDCLASSEX{
		LpfnWndProc:   syscall.NewCallback(wndProc),
		HInstance:     hInst,
		HCursor:       win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_CROSS)),
		LpszClassName: name,
	}
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	win.RegisterClassEx(&wc)
}

func createWindow(name *uint16, x, y, w, h int32) win.HWND {
	return win.CreateWindowEx(
		win.WS_EX_TOPMOST|win.WS_EX_TOOLWINDOW,
		name, nil,
		win.WS_POPUP,
		x, y, w, h,
		0, 0, win.GetModuleHandle(nil), nil,
	)
}

func show(hwnd win.HWND) {
	win.ShowWindow(hwnd, win.SW_SHOW)
	win.SetForegroundWindow(hwnd)
	win.SetFocus(hwnd)
	win.UpdateWindow(hwnd)
}

func loop() {
	var msg win.MSG
	for win.GetMessage(&msg, 0, 0, 0) > 0 {
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
	}
}

func RunSelect(path string) {
	runtime.LockOSThread()
	function.TurnDpiAware()

	img, err := loadPNG(path)
	if err != nil {
		return
	}
	rgba := toRGBA(img)
	imgW, imgH = int32(rgba.Bounds().Dx()), int32(rgba.Bounds().Dy())
	bmp = createBitmap(rgba, false)
	bmpDim = createBitmap(rgba, true)

	x, y, w, h := virtualScreen()
	name, _ := syscall.UTF16PtrFromString("SimpleQRSelect")
	registerClass(name)
	hwnd := createWindow(name, x, y, w, h)
	show(hwnd)
	loop()

	if selected {
		fmt.Printf("%d,%d,%d,%d", result.Min.X, result.Min.Y, result.Dx(), result.Dy())
	}
}
