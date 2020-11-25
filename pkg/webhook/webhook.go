package webhook

import (
	"context"
	"github.com/go-logr/logr"
	"io"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Webhook struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	server *http.Server
}

func (w *Webhook) Start(stop <-chan struct{}) error {
	idleConnsClosed := make(chan struct{})

	go func() {
		<-stop
		w.Log.Info("shutting down webhook server")

		if err := w.server.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout
			w.Log.Error(err, "error shutting down the HTTP server")
		}
		close(idleConnsClosed)
	}()

	w.setupServe()

	err := w.server.ListenAndServeTLS("", "")
	if err != nil && err != http.ErrServerClosed {
		w.Log.Error(err, "start webhook server occur error")
		return err
	}
	<-idleConnsClosed
	return nil
}

func (w *Webhook) setupServe() {

	http.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		_, _ = io.WriteString(w, "i'm ok!\n")
	})

	http.HandleFunc("/admit", w.serveAdmitHandle)
	http.HandleFunc("/mutate", w.serveMutateHandle)
	http.HandleFunc("/pods", w.serveMutateHandle)

	var c config
	server := &http.Server{
		Addr:      c.Addr,
		TLSConfig: configTLS(&c),
	}
	w.Log.Info("webhook server tls config", "config", c)
	w.server = server
}

func (wh *Webhook) serveAdmitHandle(w http.ResponseWriter, r *http.Request) {
	wh.webhookProcessor(w, r, wh.admitFuncProcessor)
}

func (wh *Webhook) serveMutateHandle(w http.ResponseWriter, r *http.Request) {
	wh.webhookProcessor(w, r, wh.mutateFuncProcessor)
}
