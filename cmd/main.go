package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rakyll/statik/fs"
	"github.com/rs/cors"
	"minivmm"
	"minivmm/api"
	_ "minivmm/statik"
)

var (
	ui      = flag.Bool("ui", false, "enable to provide ui pages")
	initNw  = flag.Bool("init-nw", false, "initialize network settings")
	resetNw = flag.Bool("reset-nw", false, "clean up network settings")
)

var (
	serverCert         = filepath.Join(os.Getenv(minivmm.EnvDir), "server.crt")
	serverKey          = filepath.Join(os.Getenv(minivmm.EnvDir), "server.key")
	listenPort         = os.Getenv(minivmm.EnvPort)
	corsAllowedOrigins = strings.Split(os.Getenv(minivmm.EnvCorsOrigins), ",")
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
		filepath.Join(os.Getenv(minivmm.EnvDir), "forwards"),
		filepath.Join(os.Getenv(minivmm.EnvDir), "images"),
		filepath.Join(os.Getenv(minivmm.EnvDir), "vms"),
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
	if listenPort == "" {
		panic("Set VMM_LISTEN_PORT")
	}

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
	c := cors.New(cors.Options{
		AllowedOrigins:   corsAllowedOrigins,
		AllowCredentials: true,
		AllowedHeaders:   []string{"authorization", "content-type"},
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PATCH"},
		Debug:            false,
	})
	handler := c.Handler(mux)

	go minivmm.ServeDHCP()
	go minivmm.UpdateIPAddress()

	log.Println("Starting minivm..")
	envNoTLS := os.Getenv(minivmm.EnvNoTLS)
	if envNoTLS == "1" || envNoTLS == "true" {
		log.Fatal(http.ListenAndServe(":"+os.Getenv(minivmm.EnvPort), handler))
	} else {
		log.Fatal(http.ListenAndServeTLS(":"+os.Getenv(minivmm.EnvPort), serverCert, serverKey, handler))
	}
}

func main() {
	flag.Parse()
	if *initNw {
		err := minivmm.InitNetns()
		if err != nil {
			panic(err)
		}
		return
	}
	if *resetNw {
		err := minivmm.ResetNetns()
		if err != nil {
			panic(err)
		}
		return
	}

	err := minivmm.StartNetwork()
	if err != nil {
		log.Println("warn network configuration error", err)
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
