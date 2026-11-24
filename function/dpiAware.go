package function

import "golang.org/x/sys/windows"

func TurnDpiAware() error {
	proc := windows.NewLazySystemDLL("user32.dll").NewProc("SetProcessDpiAwarenessContext")
	r, _, err := proc.Call(^uintptr(3)) // -4 (PER_MONITOR_AWARE_V2)
	if r == 0 {
		return err
	}
	return nil
}
