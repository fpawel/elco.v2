package main

import (
	"database/sql"
	"github.com/fpawel/elco.v2/internal/data"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

var (
	elcoDLL                        = syscall.NewLazyDLL("elco.dll")
	formPartySerialsSetVisibleProc = elcoDLL.NewProc("FormPartySerialsSetVisible")
	formPartySerialsSetCellProc    = elcoDLL.NewProc("FormPartySerialsSetCell")
	formPartySerialsSetErrorProc   = elcoDLL.NewProc("FormPartySerialsSetError")
	formPartySerialsClearErrorProc = elcoDLL.NewProc("FormPartySerialsClearError")
)

func formPartySerialsSetVisible(visible bool) {
	var v uintptr
	if visible {
		v = 1
	}
	checkSyscall(formPartySerialsSetVisibleProc.Call(v))
	if visible {
		for place := 0; place < 96; place++ {
			formPartySerialsSetCell(place%8+1, place/8+1,
				formatNullInt64(lastPartyProducts.ProductsTable().ProductAt(place).Serial))
		}
	}
}

func formPartySerialsClearError() {
	_, _, _ = formPartySerialsClearErrorProc.Call()
}

func formPartySerialsSetError(text string) {
	checkSyscall(
		formPartySerialsSetErrorProc.Call(
			uintptr(stringToUTF16Ptr(text))))
}

func formPartySerialsSetCell(col, row int, value string) {
	checkSyscall(formPartySerialsSetCellProc.Call(
		uintptr(col),
		uintptr(row),
		uintptr(stringToUTF16Ptr(value))))
}

func stringToUTF16Ptr(s string) unsafe.Pointer {
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

func StringFromUTF16Ptr(pStr unsafe.Pointer, strLen int) string {
	ptr := ptrSliceFrom(pStr, strLen)
	p := *(*[]uint16)(ptr)
	return syscall.UTF16ToString(p)
}

func formatNullInt64(v sql.NullInt64) string {
	if v.Valid {
		return strconv.FormatInt(v.Int64, 10)
	}
	return ""
}

func parseNullInt64(s string) (sql.NullInt64, error) {
	if len(s) == 0 {
		return sql.NullInt64{}, nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return sql.NullInt64{}, err
	}
	return sql.NullInt64{v, true}, nil
}

func init() {
	checkSyscall(elcoDLL.NewProc("Init").Call())

	checkSyscall(
		elcoDLL.
			NewProc("FormPartySerialsSetOnHide").
			Call(uintptr(syscall.NewCallback(func() uintptr {
				lastPartyProducts.Invalidate()
				return 0
			}))))

	checkSyscall(
		elcoDLL.
			NewProc("FormPartySerialsSetOnCellChanged").
			Call(uintptr(syscall.NewCallback(func(col, row int, pStr uintptr, strLen int) uintptr {

				place := (row-1)*8 + (col - 1)
				s := strings.TrimSpace(StringFromUTF16Ptr(unsafe.Pointer(pStr), strLen))

				showErr := func(err error) {
					formPartySerialsSetError(err.Error())
					formPartySerialsSetCell(col, row, formatNullInt64(data.GetProductSerialAtPlace(place)))
				}

				serial, err := parseNullInt64(s)

				if err != nil {
					showErr(err)
					return 0
				}

				if err = lastPartyProducts.SetProductSerialAt(place, serial); err != nil {
					showErr(err)
					return 0
				}
				formPartySerialsClearError()
				return 0
			}))))
}
