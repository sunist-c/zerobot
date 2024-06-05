package manager

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

var (
	configBytes []byte
)

func LoadPublicConfig(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	configBytes, err = io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	closeErr := f.Close()
	if closeErr != nil {
		return fmt.Errorf("failed to close file: %w", closeErr)
	}

	return nil
}

func GetJsonPublicConfig(receiver any, keys ...string) error {
	return getPublicConfig(json.Unmarshal, json.Marshal, receiver, configBytes, keys, 0)
}

func GetYamlPublicConfig(receiver any, keys ...string) error {
	return getPublicConfig(yaml.Unmarshal, yaml.Marshal, receiver, configBytes, keys, 0)
}

func getPublicConfig(unmarshaler func([]byte, any) error, marshaller func(any) ([]byte, error), receiver any, in []byte, keys []string, index int) error {
	if index == len(keys) {
		return unmarshaler(in, receiver)
	}

	currentKey := keys[index]
	buffer := map[string]any{}
	if err := unmarshaler(in, &buffer); err != nil {
		return fmt.Errorf("failed to unmarshal data at key %s: %w", currentKey, err)
	}

	object, exists := buffer[currentKey]
	if !exists {
		return fmt.Errorf("key %s not found", currentKey)
	}

	next, err := marshaller(object)
	if err != nil {
		return fmt.Errorf("failed to marshal data at key %s: %w", currentKey, err)
	}

	return getPublicConfig(unmarshaler, marshaller, receiver, next, keys, index+1)
}
