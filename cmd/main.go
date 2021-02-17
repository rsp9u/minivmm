package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rs/cors"
	"minivmm"
	"minivmm/api"
	"minivmm/ws"
)

var (
	ui      = flag.Bool("ui", false, "enable to provide ui pages")
	initNw  = flag.Bool("init-nw", false, "initialize network settings")
	resetNw = flag.Bool("reset-nw", false, "clean up network settings")
)

// DefaultedFileSystem is a file system with fallback url.
type DefaultedFileSystem struct {
	fs          http.FileSystem
	DefaultPath string
}

// Open opens the specified file. If the given file doesn't exist, it will return the default file.
func (f DefaultedFileSystem) Open(name string) (http.File, error) {
	file, err := f.fs.Open(name)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		file, err = f.fs.Open(f.DefaultPath)
		return file, err
	}
	return file, err
}

func ensureDir() error {
	dirs := []string{
		filepath.Join(minivmm.C.Dir, "forwards"),
		filepath.Join(minivmm.C.Dir, "images"),
		filepath.Join(minivmm.C.Dir, "vms"),
	}
	for _, dir := range dirs {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, encoding := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
			if strings.TrimSpace(encoding) == "gzip" {
				break
			}
			next.ServeHTTP(w, r)
			return
		}

		rec := httptest.NewRecorder()
		next.ServeHTTP(rec, r)

		var b bytes.Buffer
		gw := gzip.NewWriter(&b)
		io.Copy(gw, rec.Result().Body)
		gw.Flush()
		gw.Close()

		for k, values := range rec.Result().Header {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}
		w.Header().Add("Vary", "Accept-Encoding")
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Length", strconv.Itoa(b.Len()))
		w.WriteHeader(rec.Result().StatusCode)
		io.Copy(w, &b)
	})
}

func server() {
	mux := http.NewServeMux()
	if *ui {
		log.Println("Into ui mode..")
		defaultedFileSystem := DefaultedFileSystem{fs: http.FS(minivmm.GetAssets()), DefaultPath: "/index.html"}
		mux.Handle("/", api.AuthMiddleware(gzipMiddleware(http.StripPrefix("/", http.FileServer(defaultedFileSystem)))))
	} else {
		log.Println("Into non-ui mode..")
	}

	api.RegisterHandlers(mux)
	ws.RegisterHandlers(mux)
	c := cors.New(cors.Options{
		AllowedOrigins:   minivmm.C.CorsOrigins,
		AllowCredentials: true,
		AllowedHeaders:   []string{"authorization", "content-type"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PATCH", "OPTIONS"},
		Debug:            false,
	})
	handler := c.Handler(mux)

	go minivmm.ServeDHCP()
	go minivmm.UpdateIPAddress()

	log.Println("Starting minivm..")
	if minivmm.C.NoTLS {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", minivmm.C.Port), handler))
	} else {
		log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", minivmm.C.Port), minivmm.C.ServerCert, minivmm.C.ServerKey, handler))
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	err := minivmm.ParseConfig()
	if err != nil {
		panic(err)
	}
	minivmm.Agents, err = minivmm.InitAgentLister()
	defer minivmm.Agents.Cleanup()

	flag.Parse()
	if *initNw {
		err = minivmm.InitNetns()
		if err != nil {
			log.Println(err)
		}
		return
	}
	if *resetNw {
		err = minivmm.ResetNetns()
		if err != nil {
			panic(err)
		}
		return
	}

	err = minivmm.StartNetwork()
	if err != nil {
		log.Fatal(err)
	}
	err = ensureDir()
	if err != nil {
		log.Fatal(err)
	}
	err = minivmm.ResumeForwards()
	if err != nil {
		log.Fatal(err)
	}

	server()
}
