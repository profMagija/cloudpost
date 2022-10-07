package localrunner

import (
	_ "embed"
	"os"
	"path/filepath"
	"strings"
)

func _emit_data(rootPath string, pathAndData ...any) error {
	for i := 0; i < len(pathAndData); i += 2 {
		err := os.WriteFile(filepath.Join(rootPath, pathAndData[i].(string)), pathAndData[i+1].([]byte), 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

//go:embed python/__init__.py
var python_local_init_py []byte

//go:embed python/datastore.py
var python_local_datastore_py []byte

//go:embed python/internal.py
var python_local_internal_py []byte

//go:embed python/storage.py
var python_local_storage_py []byte

//go:embed python/pubsub.py
var python_local_pubsub_py []byte

func emit_python_local(dstPath string) error {
	err := os.MkdirAll(dstPath, 0)
	if err != nil {
		return err
	}
	return _emit_data(dstPath,
		"__init__.py", python_local_init_py,
		"datastore.py", python_local_datastore_py,
		"internal.py", python_local_internal_py,
		"storage.py", python_local_storage_py,
		"pubsub.py", python_local_pubsub_py,
	)
}

//go:embed runners/py_func_runner.py
var python_runner_func string

func emit_python_function_entry(dstPath, entry, port string) error {
	p := strings.ReplaceAll(python_runner_func, "__ENTRY__", entry)
	p = strings.ReplaceAll(p, "__PORT__", port)
	return os.WriteFile(dstPath, []byte(p), 0755)
}

//go:embed runners/py_container_runner.py
var python_runner_container string

func emit_python_container_entry(dstPath, entry, port string) error {
	p := strings.ReplaceAll(python_runner_container, "__ENTRY__", entry)
	p = strings.ReplaceAll(p, "__PORT__", port)
	return os.WriteFile(dstPath, []byte(p), 0755)
}
