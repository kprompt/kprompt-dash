package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/kprompt/kprompt-dash/internal/api"
	"github.com/kprompt/kprompt-dash/internal/kube"
	"github.com/kprompt/kprompt-dash/web"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:7474", "listen address (default localhost only)")
	contextName := flag.String("context", "", "kubeconfig context")
	open := flag.Bool("open", false, "print the UI URL (open browser yourself for now)")
	flag.Parse()

	if err := warnIfNonLocal(*addr); err != nil {
		log.Fatal(err)
	}

	clients, err := kube.Connect(*contextName)
	if err != nil {
		log.Fatalf("kubeconfig: %v", err)
	}

	mux := http.NewServeMux()
	api.Mount(mux, clients)
	mux.Handle("/", web.Handler())

	srv := &http.Server{
		Addr:              *addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		url := "http://" + *addr
		log.Printf("kprompt-dash listening on %s (context=%s)", url, clients.Context)
		if *open {
			fmt.Fprintf(os.Stderr, "Open: %s\n", url)
			_ = openBrowser(url)
		}
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func warnIfNonLocal(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		// allow :7474 form
		if strings.HasPrefix(addr, ":") {
			return fmt.Errorf("refusing bind-all address %q — use 127.0.0.1:7474 (pass --addr explicitly only if you understand the risk)", addr)
		}
		return nil
	}
	ip := net.ParseIP(host)
	if host == "localhost" || (ip != nil && ip.IsLoopback()) {
		return nil
	}
	log.Printf("WARNING: listening on non-loopback %s — dash has no auth; kube access equals this process", addr)
	return nil
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
