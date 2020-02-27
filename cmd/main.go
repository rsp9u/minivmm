package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rakyll/statik/fs"
	"github.com/rs/cors"
	"minivmm"
	"minivmm/api"
	"minivmm/ws"

	_ "minivmm/statik"
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
	if err != nil && err == os.ErrNotExist {
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

func server() {
	mux := http.NewServeMux()
	if *ui {
		log.Println("Into ui mode..")
		fs, _ := fs.New()
		defaultedFileSystem := DefaultedFileSystem{fs: fs, DefaultPath: "/index.html"}
		mux.Handle("/", api.AuthMiddleware(http.StripPrefix("/", http.FileServer(defaultedFileSystem))))
	} else {
		log.Println("Into non-ui mode..")
	}

	api.RegisterHandlers(mux)
	ws.RegisterHandlers(mux)
	c := cors.New(cors.Options{
		AllowedOrigins:   minivmm.C.CorsOrigins,
		AllowCredentials: true,
		AllowedHeaders:   []string{"authorization", "content-type"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PATCH"},
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
	err := minivmm.ParseConfig()
	if err != nil {
		panic(err)
	}

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
