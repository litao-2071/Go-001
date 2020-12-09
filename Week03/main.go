package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"golang.org/x/sync/errgroup"
)
const {
	host = "127.0.0.1"
}
type ApiConn struct {
	Port string
}

func (a *ApiConn) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(resp, "host: %s, prot: %s", req.Host, a.Port)
}

func httpServ(ctx context.Context, port string, cancel context.CancelFunc) error {
	defer cancel()
	conn := &ApiConn{Port: port}
	s := http.Server{
		Addr: host + ":" + port,
		Handler: http.Handler(conn),
	}
	go func() {
		<-ctx.Done()
		s.Shutdown(ctx)
	}()
	return s.ListenAndServe()
}

func handleSignal(ctx context.Context, cancel context.CancelFunc, c chan os.Signal) error {
	select {
	case <-ctx.Done():
		fmt.Println("cancel :", ctx.Err())
		return fmt.Errorf("cancel: %v", ctx.Err())
	case mess := <-c:
		cancel()
		return fmt.Errorf("os signal: %v", mess)
	}
}

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	ctx, cancel := context.WithCancel(context.Background())
	g, err := errgroup.WithContext(ctx)
	if err != nil {
		// handler error
	}
	g.Go(func() error {
		return httpServ(ctx, "8080", cancel)
	})
	g.Go(func() error {
		return httpServ(ctx, "8081", cancel)
	})
	g.Go(func() error {
		return handleSignal(ctx, cancel, c)
	})
	if err := g.Wait(); err != nil {
		fmt.Printf("exit:", err.Error())
	}
}
