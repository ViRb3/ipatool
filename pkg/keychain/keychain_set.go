package keychain

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	jose "github.com/dvsekhvalnov/jose2go"
)

func (k *keychain) Set(key string, data []byte) error {
	bytes, err := json.Marshal(item{
		Key:  key,
		Data: data,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	if err = k.unlock(); err != nil {
		return fmt.Errorf("failed to unlock keychain: %w", err)
	}

	token, err := jose.Encrypt(string(bytes), jose.PBES2_HS256_A128KW, jose.A256GCM, k.password,
		jose.Headers(map[string]interface{}{
			"created": time.Now().String(),
		}))
	if err != nil {
		return fmt.Errorf("failed to encode item: %w", err)
	}

	filename, err := k.filename(key)
	if err != nil {
		return fmt.Errorf("failed to resolve item path: %w", err)
	}

	err = os.WriteFile(filename, []byte(token), 0600)
	if err != nil {
		return fmt.Errorf("failed to write item: %w", err)
	}

	return nil
}
