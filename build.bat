SET APP_DIR=build
SET GOARCH=386
buildmingw32 go build -tags walk_use_cgo -o %APP_DIR%\elco.exe github.com/fpawel/elco.v2/cmd/elco
