// Copyright 2019 the orbs-network-go authors
// This file is part of the orbs-network-go library in the Orbs project.
//
// This source code is licensed under the MIT license found in the LICENSE file in the root directory of this source tree.
// The above notice should be included in all copies or substantial portions of the software.

package log

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

type httpOutputHarness struct {
	port     int
	listener net.Listener
	router   *http.ServeMux
	server   *http.Server
}

func newHttpHarness(handler http.Handler) *httpOutputHarness {
	router := http.NewServeMux()
	router.Handle("/submit-logs", handler)

	return &httpOutputHarness{
		router: router,
	}
}

type testWriter struct {
	done chan struct{}
}

func (w *testWriter) Write(p []byte) (n int, err error) {
	close(w.done)
	return len(p), nil
}

func (h *httpOutputHarness) start(t *testing.T) {
	ch := make(chan struct{})
	go func() {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err, "failed to use http port")

		h.port = listener.Addr().(*net.TCPAddr).Port
		t.Log("Serving http requests on", "127.0.0.1:%d", h.port)

		h.listener = listener

		server := &http.Server{
			Handler: h.router,
		}
		ch <- struct{}{}
		_ = server.Serve(h.listener) // no point in handling this error, it will always be an error when server dies

	}()
	<-ch

}

func (h *httpOutputHarness) stop(t *testing.T) {
	if h.server != nil {
		ctx, _ := context.WithTimeout(context.Background(), 2*time.Millisecond)
		if err := h.server.Shutdown(ctx); err != nil {
			t.Error("failed to stop http server gracefully", err)
		}
	}
}

func (h *httpOutputHarness) endpointUrl() string {
	return fmt.Sprintf("http://127.0.0.1:%d/submit-logs", h.port)
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

	w := NewHttpWriter(h.endpointUrl())
	size, err := w.Write([]byte("hello"))
	require.NoError(t, err)
	require.EqualValues(t, 5, size)

	wg.Wait()
}

func TestBulkOutput_Append(t *testing.T) {
	batch := make(chan struct{}, 1)
	timeout := make(chan struct{})
	go func() {
		time.Sleep(1 * time.Second)
		close(timeout)
	}()

	output := NewBulkOutput(
		&testWriter{batch},
		nopFormatter{},
		3)
	logger := GetLogger().WithOutput(output)

	logger.Info("Ground control to Major Tom")
	logger.Info("Commencing countdown")
	logger.Info("Engines on")
	select {
	case <-timeout:
		require.Fail(t, "Timed out waiting for batch")
	case <-batch:
	}
}

func TestBulkOutput_SetFilters(t *testing.T) {
	batch := make(chan struct{}, 1)
	timeout := make(chan struct{})
	go func() {
		time.Sleep(1 * time.Second)
		close(timeout)
	}()

	output := NewBulkOutput(
		&testWriter{batch},
		nopFormatter{},
		3)
	output.SetFilters(IgnoreMessagesMatching("Commencing"))
	logger := GetLogger().WithOutput(output)

	logger.Info("Ground control to Major Tom")
	logger.Info("Commencing countdown")
	logger.Info("Engines on")
	select {
	case <-batch:
		require.Fail(t, "Row was not filtered")
	case <-timeout:
	}
	logger.Info("Liftoff")
	<-batch
}

func TestBulkOutput_Append_Http(t *testing.T) {
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
		NewBulkOutput(
			NewHttpWriter(h.endpointUrl()),
			NewJsonFormatter(),
			3))

	logger.Info("Ground control to Major Tom")
	logger.Info("Commencing countdown")
	logger.Info("Engines on")

	wg.Wait()
}

func TestHttpWriter_ConsidersResponse201AsSuccess(t *testing.T) {
	h := newHttpHarness(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		_, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)
		w.WriteHeader(201)
	}))
	h.start(t)
	defer h.stop(t)

	httpWriter := NewHttpWriter(h.endpointUrl())
	_, err := httpWriter.Write([]byte("{}"))

	require.NoError(t, err, "HttpWriter considered HTTP 201 as an error response")
}

func TestHttpWriter_RequestFailsWhenTimeoutReached(t *testing.T) {
	h := newHttpHarness(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		_, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)
		time.Sleep(time.Second)
		w.WriteHeader(200)
	}))
	h.start(t)
	defer h.stop(t)

	httpWriter := NewHttpWriterWithTimeout(h.endpointUrl(), 100*time.Millisecond)
	_, err := httpWriter.Write([]byte("{}"))

	require.Error(t, err, "HTTP request did not fail")
	require.Regexp(t, "Timeout", err.Error(), "HTTP request failed with a non timeout related error")
}
