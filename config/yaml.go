package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func yaml_reEncode(in any, out any) error {
	data, err := yaml.Marshal(in)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, out)
	return err
}

func (f *Flock) UnmarshalYAML(value *yaml.Node) error {
	m := make(map[string]any)

	err := yaml_reEncode(value, m)

	if err != nil {
		return fmt.Errorf("invalid flock specification: %w", err)
	}

	f.Config = make(map[string]map[string]any)
	conf_map, ok := m["config"]
	if ok {
		err = yaml_reEncode(conf_map, f.Config)
		if err != nil {
			return fmt.Errorf("invalid flock configuration: %w", err)
		}
	}

	f.Components = make([]Component, 0)
	components := make([]map[string]any, 0)
	comp_list, ok := m["components"]
	if ok {
		err = yaml_reEncode(comp_list, &components)
		if err != nil {
			return fmt.Errorf("invalid flock component spec: %w", err)
		}

		for _, component := range components {
			switch component["type"] {
			case nil:
				return fmt.Errorf("invalid flock component spec: must have type")
			case "function":
				res := new(Function)
				err = yaml_reEncode(component, res)
				if err != nil {
					return fmt.Errorf("invalid flock component spec: %w", err)
				}
				f.Components = append(f.Components, res)
			case "container":
				res := new(Container)
				err = yaml_reEncode(component, res)
				if err != nil {
					return fmt.Errorf("invalid flock component spec: %w", err)
				}
				f.Components = append(f.Components, res)
			default:
				return fmt.Errorf("invalid flock component spec: unknown type '%v'", component["type"])
			}
		}
	}

	return nil
}

func (f *FileSpec) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		f.Src = value.Value
		f.Dst = value.Value
		return nil
	case yaml.SequenceNode:
		switch len(value.Content) {
		case 1:
			yaml_reEncode(value.Content[0], &f.Src)
			f.Dst = f.Src
			return nil
		case 2:
			yaml_reEncode(value.Content[0], &f.Src)
			yaml_reEncode(value.Content[1], &f.Dst)
			return nil
		default:
			return fmt.Errorf("invalid file specification: expected 1 or 2 elements, got %d", len(value.Content))
		}
	default:
		return fmt.Errorf("invalid file specification: expected string or list, got %v", value.Kind)
	}
}
