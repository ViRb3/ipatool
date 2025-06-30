package keychain

import (
	"encoding/json"
	"fmt"
	"os"

	jose "github.com/dvsekhvalnov/jose2go"
)

func (k *keychain) Get(key string) ([]byte, error) {
	filename, err := k.filename(key)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve item path: %w", err)
	}

	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read item: %w", err)
	}

	if err = k.unlock(); err != nil {
		return nil, fmt.Errorf("failed to unlock keychain: %w", err)
	}

	payload, _, err := jose.Decode(string(bytes), k.password)
	if err != nil {
		return nil, fmt.Errorf("failed to decode item: %w", err)
	}

	var item item
	err = json.Unmarshal([]byte(payload), &item)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal item: %w", err)
	}

	return item.Data, nil
}
