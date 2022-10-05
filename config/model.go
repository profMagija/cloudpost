package config

import (
	"fmt"
	"reflect"
)

type (
	Flock struct {
		Components []Component
		Config     map[string]map[string]any

		Root string `yaml:"-"`
	}

	ComponentKind int

	Component interface {
		GetName() string
		GetType() ComponentKind
	}

	Function struct {
		Name  string
		Files []FileSpec
		Entry string

		TriggerTopic string
	}

	Container struct {
		Name  string
		Files []FileSpec
		Entry string

		TriggerTopic string
		TriggerPath  string
	}

	FileSpec struct {
		Src string
		Dst string
	}
)

const (
	ComponentKind_Service ComponentKind = iota
)

func (s *Function) GetName() string {
	return s.Name
}

func (s *Function) GetType() ComponentKind {
	return ComponentKind_Service
}

func (s *Container) GetName() string {
	return s.Name
}

func (s *Container) GetType() ComponentKind {
	return ComponentKind_Service
}

func _merge(def, override any) (any, error) {
	if def == nil {
		return override, nil
	}
	if override == nil {
		return def, nil
	}

	var err error

	dv := reflect.ValueOf(def)
	ov := reflect.ValueOf(override)
	if dv.Kind() == reflect.Map && ov.Kind() == reflect.Map {
		result := make(map[string]any)

		dIter := dv.MapRange()
		for dIter.Next() {
			key := fmt.Sprint(dIter.Key().Interface())
			result[key] = dIter.Value().Interface()
		}

		oIter := ov.MapRange()
		for oIter.Next() {
			key := fmt.Sprint(oIter.Key().Interface())
			oValue := oIter.Value().Interface()
			if result[key] != nil {
				result[key] = oValue
			} else {
				result[key], err = _merge(result[key], oValue)
				if err != nil {
					return nil, err
				}
			}
		}
		return result, nil
	} else if dv.Kind() == ov.Kind() {
		return ov.Interface(), nil
	} else {
		return nil, fmt.Errorf("cannot override %s with %s", dv.Kind(), ov.Kind())
	}
}

func (s *Flock) ResolveConfig(cfg string) (map[string]any, error) {
	def := s.Config["default"]
	override := s.Config[cfg]

	res, err := _merge(def, override)
	return res.(map[string]any), err
}
