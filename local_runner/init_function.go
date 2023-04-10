package localrunner

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/profMagija/cloudpost/config"

	"github.com/gorilla/mux"
)

func cloudfunction_copyFiles(td string, root string, f *config.Function, sourceMap map[string]string, port int) error {
	for _, file := range f.Files {
		err := _copyFiles(filepath.Join(root, file.Src), file.Dst, td, sourceMap)
		if err != nil {
			return err
		}
	}

	err := emit_python_local(filepath.Join(td, "cloud"))
	if err != nil {
		return err
	}

	entry := f.Entry
	if entry == "" {
		entry = "main"
	}

	err = emit_python_function_entry(filepath.Join(td, "__entry.py"), entry, strconv.Itoa(port))
	if err != nil {
		return err
	}

	return nil
}

func cloudfunction_run(flock *config.Flock, f *config.Function, port int, env map[string]string, r *mux.Router, startWg *sync.WaitGroup) {
	if f.TriggerTopic != "" {
		register_queue_target(f.TriggerTopic, fmt.Sprintf("http://localhost:%d/event", port))
	}

	name := f.Name

	prefix := "/" + f.Name + "/"
	r.PathPrefix(prefix).Methods("POST").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forward_request(prefix, fmt.Sprintf("http://localhost:%d/", port), w, r)
	})

	sourceMap := make(map[string]string)

	root := flock.Root

	td, err := os.MkdirTemp(os.TempDir(), "localrunner-")
	defer os.RemoveAll(td)

	if err != nil {
		local_log_error("[%s] error creating temp dir: %v", name, err)
		return
	}

	startWg.Done()

	err = withReloader(f.Name, sourceMap, func() error {
		err = cloudfunction_copyFiles(td, root, f, sourceMap, port)
		if err != nil {
			local_log_error("[%s] error copying function files: %v", name, err)
			return err
		}

		return nil
	}, func(interrupt chan error) (func() error, error) {
		cmd := exec.Command("python", "-u", "__entry.py")
		cmd.Dir = td
		cmd.Stdout = &namedPrinter{color: port%10 + 30, name: f.Name}
		cmd.Stderr = &namedPrinter{color: port%10 + 30, name: f.Name}

		envList := []string{
			"LOCALRUNNER_ADDR=http://127.0.0.1:5000",
			"FLASK_DEBUG=1",
		}
		envList = append(envList, os.Environ()...)
		for k, v := range env {
			envList = append(envList, fmt.Sprintf("%s=%s", k, v))
		}
		cmd.Env = envList

		err = cmd.Start()
		if err != nil {
			local_log_error("[%s] error function process: %v", name, err)
			return nil, err
		}

		go func() {
			interrupt <- cmd.Wait()
		}()

		return func() error {
			local_log_verbose("[%s] killing", name)
			err := cmd.Process.Kill()
			if err != nil {
				return err
			}
			cmd.Wait()
			return nil
		}, nil
	})

	if err != nil {
		local_log_error("[%s] error: %v", name, err)
	}
}
