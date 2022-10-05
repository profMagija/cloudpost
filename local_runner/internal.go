package localrunner

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

var queues = make(map[string][]string)

func register_queue_target(queue, url string) {
	queues[queue] = append(queues[queue], url)
}

func internal_QueuePublish(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)

	name := mux.Vars(r)["name"]

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	realData := []byte(`{"message":{"data":"`)
	realData = append(realData, []byte(base64.StdEncoding.EncodeToString(data))...)
	realData = append(realData, []byte(`"}}`)...)

	log.Printf(" publishing to queue %s: %s", name, string(realData))

	for _, dest := range queues[name] {
		go func(dest string) {
			r, err := http.Post(dest, "application/json", bytes.NewBuffer(realData))
			if err != nil {
				log.Printf("error sending to subscriber: %v", err)
			}
			if r.StatusCode >= 400 {
				reason, err := io.ReadAll(r.Body)
				if err != nil {
					reason = []byte("<error while reading>")
				}
				log.Printf("subscriber errored: %s", string(reason))
			}
		}(dest)
	}

	http.Error(w, "ok", 200)
}

// ---------------------------------------------

var datastore = make(map[string]map[string]map[any]map[string]any)

func datastore_make_key(key string) any {
	i, err := strconv.Atoi(key)
	if err != nil {
		return key
	} else {
		return i
	}
}

func datastore_get_entity(namespace, kind string, key any) map[string]any {
	ns := datastore[namespace]
	if ns != nil {
		kd := ns[kind]
		if kd != nil {
			return kd[key]
		}
	}
	return nil
}

func datastore_delete_entity(namespace, kind string, key any) {
	ns := datastore[namespace]
	if ns != nil {
		kd := ns[kind]
		if kd != nil {
			delete(kd, key)
		}
	}
}

func datastore_put_entity(namespace, kind string, key any, entity map[string]any) {
	ns := datastore[namespace]
	if ns == nil {
		ns = make(map[string]map[any]map[string]any)
		datastore[namespace] = ns
	}

	kd := ns[kind]
	if kd == nil {
		kd = make(map[any]map[string]any)
		ns[kind] = kd
	}

	entity["#key"] = key
	kd[key] = entity
}

func internal_DatastoreListNamespace(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "only GET", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	namespace := vars["namespace"]
	kind := vars["kind"]

	fmt.Printf(" [\x1b[33m~\x1b[m] datastore / list : %s/%s\n", namespace, kind)

	ns := datastore[namespace]
	result := make([]any, 0)
	if ns != nil {
		for _, value := range ns[kind] {
			result = append(result, value)
		}
	}

	data, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Add("content-type", "application/json")
	w.Header().Add("content-length", strconv.Itoa(len(data)))
	w.Write(data)
}

func internal_DatastoreEntity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	kind := vars["kind"]
	key := datastore_make_key(vars["key"])

	fmt.Printf(" [\x1b[33m~\x1b[m] datastore / %s : %s/%s/%s\n", strings.ToLower(r.Method), namespace, kind, key)

	switch r.Method {
	case http.MethodGet:
		ent := datastore_get_entity(namespace, kind, key)
		data, err := json.Marshal(ent)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Add("content-type", "application/json")
		w.Header().Add("content-length", strconv.Itoa(len(data)))
		w.Write(data)
		return
	case http.MethodPut:
		ent := make(map[string]any)
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		err = json.Unmarshal(data, &ent)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		dataResp, err := json.Marshal(ent)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		datastore_put_entity(namespace, kind, key, ent)

		w.Header().Add("content-type", "application/json")
		w.Header().Add("content-length", strconv.Itoa(len(dataResp)))
		w.Write(dataResp)
		return
	case http.MethodDelete:
		datastore_delete_entity(namespace, kind, key)
		http.Error(w, "ok", 200)
		return
	}
}

type storageObject struct {
	mimeType string
	data     []byte
}

var storage = make(map[string]map[string]storageObject)

func internal_StorageListBucket(w http.ResponseWriter, r *http.Request) {
	bucket := mux.Vars(r)["bucket"]

	res := make([]string, 0, len(storage[bucket]))
	for k := range storage[bucket] {
		res = append(res, k)
	}

	data, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	w.Header().Add("content-type", "application/json")
	w.Header().Add("content-length", strconv.Itoa(len(data)))
	w.Write(data)
}

func internal_StorageObject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := strings.ToLower(vars["bucket"])
	path := strings.ToLower(vars["path"])

	bkt := storage[bucket]
	if bkt == nil {
		http.Error(w, "not found", 404)
		return
	}

	switch r.Method {
	case http.MethodGet:
		obj, ok := bkt[path]
		if !ok {
			http.Error(w, "not found", 404)
			return
		}
		w.Header().Add("content-type", obj.mimeType)
		w.Header().Add("content-length", strconv.Itoa(len(obj.data)))
		w.Write(obj.data)
		return
	case http.MethodPut:
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		mimeType := "application/octet-stream"
		if ct, ok := r.Header["Content-Type"]; ok {
			mimeType = ct[0]
		}

		bkt[path] = storageObject{
			mimeType: mimeType,
			data:     data,
		}

		http.Error(w, "ok", 200)
		return
	case http.MethodDelete:
		delete(bkt, path)
		http.Error(w, "ok", 200)
		return
	}
}

func registerInternalHandlers(r *mux.Router) {
	r.Path("/_internal/queue/{name}/publish").Methods("GET").HandlerFunc(internal_QueuePublish)
	r.Path("/_internal/datastore/{namespace}/{kind}").Methods("GET").HandlerFunc(internal_DatastoreListNamespace)
	r.Path("/_internal/datastore/{namespace}/{kind}/{key}").Methods("GET", "POST", "DELETE").HandlerFunc(internal_DatastoreEntity)
	r.Path("/_internal/storage/{bucket}").Methods("GET").HandlerFunc(internal_StorageListBucket)
	r.Path("/_internal/storage/{bucket}/{path:.*}").Methods("GET", "PUT", "DELETE").HandlerFunc(internal_StorageObject)
}
