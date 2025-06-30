package keychain

import (
	"fmt"
	"os"
)

func (k *keychain) Remove(key string) error {
	filename, err := k.filename(key)
	if err != nil {
		return fmt.Errorf("failed to resolve item path: %w", err)
	}

	err = os.Remove(filename)
	if err != nil {
		return fmt.Errorf("failed to remove item: %w", err)
	}

	return nil
}
