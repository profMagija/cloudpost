package localrunner

import (
	"cloudpost/config"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type seedData struct {
	Datastore map[string]map[string]map[any]map[string]any
}

func do_seed(flock *config.Flock) {
	data, err := os.ReadFile(filepath.Join(flock.Root, "cloudpost-seed.yml"))
	if err != nil {
		return
	}

	sd := new(seedData)

	err = yaml.Unmarshal(data, sd)
	if err != nil {
		log.Printf("error seeding: %v", err)
		return
	}

	entities := 0

	for namespace, ns := range sd.Datastore {
		for kind, kd := range ns {
			for key, entity := range kd {
				if sKey, ok := key.(string); ok {
					key = datastore_make_key(sKey)
				}
				datastore_put_entity(namespace, kind, key, entity)
				entities += 1
			}
		}
	}

	if entities != 0 {
		fmt.Printf(" [\x1b[32m*\x1b[m] seeded %d entities\n", entities)
	}
}
