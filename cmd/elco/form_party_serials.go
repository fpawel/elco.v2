package main

import (
	"fmt"
	"reflect"
	"syscall"
	"unsafe"
)

var (
	elcoDLL                     = syscall.NewLazyDLL("elco.dll")
	formPartySerialsShowProc    = elcoDLL.NewProc("FormPartySerialsShow")
	formPartySerialsSetCellProc = elcoDLL.NewProc("FormPartySerialsSetCell")
)

func formPartySerialsShow() {
	if r, _, err := formPartySerialsShowProc.Call(); r == 0 {
		panic(err)
	}
}

func formPartySerialsSetCell(col, row int, value string) {
	if r, _, err := formPartySerialsSetCellProc.Call(
		uintptr(col),
		uintptr(row),
		uintptr(utf16StringPtr(value))); r == 0 {
		panic(err)
	}
}

func utf16StringPtr(s string) unsafe.Pointer {
	p, err := syscall.UTF16PtrFromString(s)
	if err == syscall.EINVAL {
		p, err = syscall.UTF16PtrFromString("")
	}
	if err != nil {
		panic(err)
	}
	return unsafe.Pointer(p)
}

func checkSyscall(r1 uintptr, _ uintptr, lastErr error) {
	if r1 == 0 {
		panic(lastErr)
	}
}

func ptrSliceFrom(p unsafe.Pointer, s int) unsafe.Pointer {
	return unsafe.Pointer(&reflect.SliceHeader{Data: uintptr(p), Len: s, Cap: s})
}

func init() {
	checkSyscall(elcoDLL.NewProc("Init").Call())

	checkSyscall(
		elcoDLL.
			NewProc("FormPartySerialsSetOnCellChanged").
			Call(uintptr(syscall.NewCallback(func(col, row int, pStr uintptr, strLen int) uintptr {
				ptr := ptrSliceFrom(unsafe.Pointer(pStr), strLen)
				p := *(*[]uint16)(ptr)
				fmt.Println(col, row, "\""+syscall.UTF16ToString(p)+"\"")

				return 0

			}))))
}
