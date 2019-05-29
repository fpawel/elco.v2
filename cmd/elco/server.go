package main

import (
	"context"
	"github.com/fpawel/elco.v2/internal/api"
	"github.com/powerman/rpc-codec/jsonrpc2"
	"golang.org/x/sys/windows/registry"
	"io"
	"net"
	"net/http"
	"net/rpc"
)

func startHttpServer() func() {

	// Server export an object of type ExampleSvc.
	for _, x := range []interface{}{
		&api.LastPartySvc{},
		&api.EccInfoSvc{},
	} {
		if err := rpc.Register(x); err != nil {
			panic(err)
		}
	}

	// Server provide a HTTP transport on /rpc endpoint.
	http.Handle("/rpc", jsonrpc2.HTTPHandler(nil))

	srv := new(http.Server)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "hello world\n")
	})
	lnHTTP, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	addr := "http://" + lnHTTP.Addr().String()
	log.Info(addr)
	key, _, err := registry.CreateKey(registry.CURRENT_USER, `elco\http`, registry.ALL_ACCESS)
	if err != nil {
		panic(err)
	}
	if err := key.SetStringValue("addr", addr); err != nil {
		panic(err)
	}
	log.ErrIfFail(key.Close)

	go func() {
		if err := srv.Serve(lnHTTP); err != http.ErrServerClosed {
			log.PrintErr(err)
		}
		_ = lnHTTP.Close()
	}()

	return func() {
		if err := srv.Shutdown(context.TODO()); err != nil {
			log.PrintErr(err)
		}
	}
}
