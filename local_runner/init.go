package localrunner

import (
	"cloudpost/config"
	_ "embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/otiai10/copy"
)

func Init(flock *config.Flock) {
	os.Setenv("GCP_PROJECT", "CP_TEST")
	os.Setenv("GCLOUD_PROJECT", "CP_TEST")

	r := mux.NewRouter()

	registerInternalHandlers(r)

	env := make(map[string]string)

	cfg, err := flock.ResolveConfig("local")
	if err != nil {
		panic(err)
	}

	for k, v := range cfg {
		if strings.HasPrefix(k, "env:") {
			env[strings.TrimPrefix(k, "env:")] = fmt.Sprint(v)
		}
	}

	startWg := new(sync.WaitGroup)

	port := 6000
	for _, component := range flock.Components {
		switch c := component.(type) {
		case *config.Function:
			startWg.Add(1)
			go cloudfunction_run(flock, c, port, env, r, startWg)
			port += 1
		case *config.Container:
			startWg.Add(1)
			go container_create_app(flock, c, port, env, r, startWg)
			port += 1
		}
	}

	go do_seed(flock)

	go func() {
		startWg.Wait()
		fmt.Println(" [\x1b[32m*\x1b[m] all services starting")
	}()

	http.ListenAndServe("localhost:5000", r)
}

func _cp(src, dst string) error {
	err := os.MkdirAll(filepath.Dir(dst), 0)
	if err != nil {
		return err
	}

	info, err := os.Stat(dst)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if err == nil && info.IsDir() {
		err = os.MkdirAll(dst, 0)
		if err != nil {
			return err
		}

		dst = filepath.Join(dst, filepath.Base(src))
	}

	return copy.Copy(src, dst)
}

func _copyFiles(srcPath, dstPath, td string, sourceMap map[string]string) error {
	if strings.Contains(srcPath, "*") {
		files, err := filepath.Glob(srcPath)
		if err != nil {
			return err
		}
		for _, sp := range files {
			dst := filepath.Join(td, dstPath)
			err := os.MkdirAll(filepath.Dir(dst), 0)
			if err != nil {
				return err
			}
			err = _cp(sp, dst)
			if err != nil {
				return err
			}
			sourceMap[dst] = sp
		}
	} else {
		dst := filepath.Join(td, dstPath)
		err := _cp(srcPath, dst)
		if err != nil {
			return err
		}
		sourceMap[dst] = srcPath
	}
	return nil
}

var printerLock sync.Mutex

type namedPrinter struct {
	color int
	name  string
	buf   []byte
}

// Write implements io.Writer
func (p *namedPrinter) Write(data []byte) (n int, err error) {
	printerLock.Lock()
	defer printerLock.Unlock()

	startI := len(p.buf)
	p.buf = append(p.buf, data...)
	lastI := 0
	for i := startI; i < len(p.buf); i++ {
		if p.buf[i] == '\n' {
			if len(p.name) > 20 {
				p.name = p.name[:17] + "..."
			}
			fmt.Printf("\x1b[%dm%20s | %s\x1b[m\n", p.color, p.name, p.buf[lastI:i])
			lastI = i + 1
			i++
		}
	}

	p.buf = p.buf[lastI:]
	return len(data), nil
}
