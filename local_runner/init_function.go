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
	"sync"

	"github.com/gorilla/mux"
)

func cloudfunction_run(flock *config.Flock, f *config.Function, port int, env map[string]string, r *mux.Router, startWg *sync.WaitGroup) {
	l := log.New(os.Stdout, "["+f.Name+"] ", log.Lmsgprefix)

	if f.TriggerTopic != "" {
		register_queue_target(f.TriggerTopic, fmt.Sprintf("http://localhost:%d/event", port))
	}

	prefix := "/" + f.Name + "/"
	r.PathPrefix(prefix).Methods("POST").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forward_request(prefix, fmt.Sprintf("http://localhost:%d/", port), w, r)
	})

	td, err := os.MkdirTemp(os.TempDir(), "localrunner-")
	if err != nil {
		l.Printf("error making temp dir: %v", err)
		return
	}
	defer os.RemoveAll(td)

	sourceMap := make(map[string]string)

	root := flock.Root

	for _, file := range f.Files {
		err := _copyFiles(filepath.Join(root, file.Src), file.Dst, td, sourceMap)
		if err != nil {
			l.Printf("error copying function files: %v", err)
			return
		}
	}

	err = emit_python_local(filepath.Join(td, "cloud"))
	if err != nil {
		l.Printf("error copying function files: %v", err)
		return
	}

	entry := f.Entry
	if entry == "" {
		entry = "main"
	}

	err = emit_python_function_entry(filepath.Join(td, "__entry.py"), entry, strconv.Itoa(port))
	if err != nil {
		l.Printf("error copying function files: %v", err)
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
