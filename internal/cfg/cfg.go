package cfg

import "github.com/lxn/walk"

const (
	ComportKey    = "Comport"
	ComportGasKey = "ComportGas"
)

func Str(key string) string {

	s, _ := walk.App().Settings().Get(key)
	return s
}

func PutStr(key, value string) {
	if err := walk.App().Settings().Put(key, value); err != nil {
		panic(err)
	}
}
