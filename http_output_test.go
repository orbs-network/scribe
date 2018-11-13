package log

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

type httpOutputHarness struct {
	port     uint16
	listener net.Listener
	router   *http.ServeMux
	server   *http.Server
}

func newHttpHarness(handler http.Handler) *httpOutputHarness {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	port := uint16(20000 + r.Intn(40000))

	router := http.NewServeMux()
	router.Handle("/submit-logs", handler)

	return &httpOutputHarness{
		port:   port,
		router: router,
	}
}

func (h *httpOutputHarness) start(t *testing.T) {
	go func() {
		address := fmt.Sprintf("0.0.0.0:%d", h.port)
		t.Log("Serving http requests on", address)

		listener, err := net.Listen("tcp", address)
		h.listener = listener

		require.NoError(t, err, "failed to use http port")

		server := &http.Server{
			Handler: h.router,
		}
		err = server.Serve(h.listener)
		require.NoError(t, err, "failed to serve http requests")
	}()
}

func (h *httpOutputHarness) stop(t *testing.T) {
	if h.server != nil {
		ctx, _ := context.WithTimeout(context.Background(), 2*time.Millisecond)
		if err := h.server.Shutdown(ctx); err != nil {
			t.Error("failed to stop http server gracefully", err)
		}
	}
}

func TestHttpWriter_Write(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	h := newHttpHarness(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		require.EqualValues(t, []byte("hello"), body)

		w.WriteHeader(200)
		wg.Done()
	}))
	h.start(t)
	defer h.stop(t)

	w := NewHttpWriter(fmt.Sprintf("http://localhost:%d/submit-logs", h.port))
	size, err := w.Write([]byte("hello"))
	require.NoError(t, err)
	require.EqualValues(t, 5, size)

	wg.Wait()
}

func TestHttpOutput_Append(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	h := newHttpHarness(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		lines := strings.Split(string(body), "\n")
		require.EqualValues(t, 4, len(lines))

		w.WriteHeader(200)
		wg.Done()
	}))
	h.start(t)
	defer h.stop(t)

	logger := GetLogger().WithOutput(
		NewHttpOutput(
			NewHttpWriter(fmt.Sprintf("http://localhost:%d/submit-logs", h.port)),
			NewJsonFormatter(),
			3))

	logger.Info("Ground control to Major Tom")
	logger.Info("Commencing countdown")
	logger.Info("Engines on")

	wg.Wait()
}