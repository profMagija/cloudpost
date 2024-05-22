package localrunner

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/profMagija/cloudpost/config"

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

	td, err := os.MkdirTemp(os.TempDir(), "localrunner-")
	if err != nil {
		local_log_error("[%s] error making temp dir: %v", f.Name, err)
		return
	}
	defer os.RemoveAll(td)

	sourceMap := make(map[string]string)

	startWg.Done()

	err = withReloader(f.Name, sourceMap, func() error {
		dfPath := filepath.Join(flock.Root, "docker", "Dockerfile."+f.Name)

		data, err := os.ReadFile(dfPath)
		if err != nil {
			return err
		}

		sourceMap["<dockerfile>"] = dfPath

		for _, line := range strings.Split(string(data), "\n") {
			lineSplit := strings.Fields(line)
			if len(lineSplit) > 0 && lineSplit[0] == "COPY" {
				err = _copyFiles(filepath.Join(flock.Root, lineSplit[1]), lineSplit[2], td, sourceMap)
				if err != nil {
					return err
				}
			}
		}

		err = emit_python_local(filepath.Join(td, "cloud"))
		if err != nil {
			return err
		}

		err = emit_python_container_entry(filepath.Join(td, "__entry.py"), f.Entry, strconv.Itoa(port), f.IsNative)
		if err != nil {
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

		go func() {
			interrupt <- cmd.Wait()
		}()

		return func() error {
			err = cmd.Process.Kill()
			if err != nil {
				return err
			}
			cmd.Wait()
			return nil
		}, nil
	})

	if err != nil {
		local_log_error("[%s] error running process: %v", f.Name, err)
	}
}
