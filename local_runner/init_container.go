package localrunner

import (
	"cloudpost/config"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/mux"
)

func container_proxyHandler(w http.ResponseWriter, r *http.Request, prefix string, port int) {
	forward_request(prefix, fmt.Sprintf("http://localhost:%d/", port), w, r)
}

func container_create_app(flock *config.Flock, f *config.Container, port int, env map[string]string, r *mux.Router, startWg *sync.WaitGroup) {
	prefix := "/" + f.Name + "/"

	r.PathPrefix(prefix).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		container_proxyHandler(w, r, prefix, port)
	})

	if f.TriggerTopic != "" {
		register_queue_target(
			f.TriggerTopic,
			fmt.Sprintf("http://localhost:%d%s", port, f.TriggerPath),
		)
	}

	l := log.New(os.Stdout, "["+f.Name+"] ", log.Lmsgprefix)

	td, err := os.MkdirTemp(os.TempDir(), "localrunner-")
	if err != nil {
		l.Printf("error making temp dir: %v", err)
		return
	}
	// defer os.RemoveAll(td)

	sourceMap := make(map[string]string)

	data, err := os.ReadFile(filepath.Join(flock.Root, "docker", "Dockerfile."+f.Name))
	if err != nil {
		l.Printf("error reading dockerfile: %v", err)
		return
	}

	for _, line := range strings.Split(string(data), "\n") {
		lineSplit := strings.Fields(line)
		if len(lineSplit) > 0 && lineSplit[0] == "COPY" {
			err = _copyFiles(filepath.Join(flock.Root, lineSplit[1]), lineSplit[2], td, sourceMap)
			if err != nil {
				l.Printf("error copying files: %v", err)
				return
			}
		}
	}

	err = emit_python_local(filepath.Join(td, "cloud"))
	if err != nil {
		l.Printf("error copying container files: %v", err)
		return
	}

	err = emit_python_container_entry(filepath.Join(td, "__entry.py"), f.Entry, strconv.Itoa(port))
	if err != nil {
		l.Printf("error copying container files: %v", err)
		return
	}

	cmd := exec.Command("python", "-u", "__entry.py")
	cmd.Dir = td
	cmd.Stdout = &namedPrinter{color: port%10 + 30, name: f.Name}
	cmd.Stderr = &namedPrinter{color: port%10 + 30, name: f.Name}

	envList := []string{
		"LOCALRUNNER_ADDR=http://localhost:5000",
		"FLASK_ENV=development",
	}
	envList = append(envList, os.Environ()...)
	for k, v := range env {
		envList = append(envList, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = envList

	startWg.Done()

	err = cmd.Run()
	if err != nil {
		l.Printf("error function process: %v", err)
		return
	}
}
