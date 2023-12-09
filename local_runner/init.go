package localrunner

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/profMagija/cloudpost/config"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/mux"
	"github.com/otiai10/copy"
)

var components_Functions []string
var components_Containers []string

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
			components_Functions = append(components_Functions, c.Name)
			go cloudfunction_run(flock, c, port, env, r, startWg)
			port += 1
		case *config.Container:
			startWg.Add(1)
			components_Containers = append(components_Containers, c.Name)
			go container_create_app(flock, c, port, env, r, startWg)
			port += 1
		case *config.Bucket:
			s := make(map[string]*storageObject)
			storage[c.Name] = s
			init_bucket(flock, c, s)
		}
	}

	go do_seed(flock)

	go func() {
		startWg.Wait()
		local_log_success("all services starting")
	}()

	err = http.ListenAndServe("localhost:5000", r)
	panic(err)
}

func _cp(src, dst string) error {
	err := os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return err
	}

	info, err := os.Stat(dst)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if err == nil && info.IsDir() {
		err = os.MkdirAll(dst, 0755)
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
			err := os.MkdirAll(filepath.Dir(dst), 0755)
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

func removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func withReloader(name string, sourceMap map[string]string, setup func() error, starter func(interrupt chan error) (func() error, error)) error {
	for {
		var paths []string

		err := setup()
		if err != nil {
			return err
		}

		for _, p := range sourceMap {
			fi, err := os.Stat(p)
			if err != nil {
				return err
			}

			if fi.IsDir() {
				filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						return err
					}

					if d.Name() == "__pycache__" || d.Name() == ".git" {
						return filepath.SkipDir
					}

					paths = append(paths, path)
					return nil
				})
			} else {
				paths = append(paths, p)
			}
		}

		paths = removeDuplicateStr(paths)

		interrupt := make(chan error, 1)

		local_log_verbose("[%s] %v", name, paths)

		if len(paths) == 0 {
			_, err := starter(interrupt)
			if err != nil {
				return err
			}
		}

		rw, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}

		for _, p := range paths {
			err := rw.Add(p)
			if err != nil {
				return err
			}
		}

		local_log_verbose("[%s] starting", name)

		ender, err := starter(interrupt)
		if err != nil {
			return err
		}

		local_log_verbose("[%s] started", name)

		select {
		case e := <-rw.Events:
			rw.Close()

			local_log_verbose("[%s] hot reloading (%v)", name, e)
			err := ender()
			if err != nil {
				return err
			}
			time.Sleep(100 * time.Millisecond)
		case err := <-rw.Errors:
			local_log_verbose("[%s] hot erroring (%v)", name, err)
			ender()
			rw.Close()
			return err
		case err := <-interrupt:
			local_log_verbose("[%s] hot quitting (%v)", name, err)
			rw.Close()
			return err
		}

	}
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
			text := [][]byte{p.buf[lastI:i]}
			if bytes.HasPrefix(text[0], []byte("{\"")) {
				text = p.formatJson(text[0])
			}

			for _, line := range text {
				fmt.Printf("\x1b[%dm%20s | %s\x1b[m\n", p.color, p.name, line)
			}
			lastI = i + 1
			i++
		}
	}

	p.buf = p.buf[lastI:]
	return len(data), nil
}

func (p *namedPrinter) formatJson(origText []byte) [][]byte {
	var log structLog
	err := json.Unmarshal(origText, &log)
	if err != nil {
		return [][]byte{origText}
	}

	str := fmt.Sprintf("%-7s %s", optstr(log.Severity, "DEFAULT"), optstr(log.Message, ""))

	if log.SourceLocation != nil {
		pathstr := optstr(log.SourceLocation.File, "") + ":" + optstr(log.SourceLocation.Function, "")

		str = fmt.Sprintf("%-20s | %s", pathstr, str)
	}

	if log.Fields != nil {
		for k, v := range log.Fields {
			vSer, err := json.MarshalIndent(v, "        ", "  ")
			if err != nil {
				vSer = []byte(fmt.Sprintf("%v", v))
			}
			str += fmt.Sprintf("\n    %s = %s", k, string(vSer))
		}
	}

	if log.Exception != nil {
		str += "\n" + *log.Exception
	}

	var res [][]byte

	for _, line := range strings.Split(str, "\n") {
		res = append(res, []byte(line))
	}

	return res
}

func optstr(opt *string, def string) string {
	if opt == nil {
		return def
	}
	return *opt
}

type structLog struct {
	Severity       *string `json:"severity"`
	SourceLocation *struct {
		File     *string `json:"file"`
		Function *string `json:"function"`
	} `json:"logging.googleapis.com/sourceLocation"`
	Message   *string        `json:"message"`
	Fields    map[string]any `json:"fields"`
	Exception *string        `json:"exception"`
}
