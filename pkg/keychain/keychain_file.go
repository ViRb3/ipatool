package keychain

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type item struct {
	Key  string
	Data []byte
}

func (k *keychain) filename(key string) (string, error) {
	directory, err := k.resolveDirectory()
	if err != nil {
		return "", err
	}

	return filepath.Join(directory, strings.ReplaceAll(key, "/", "%2F")), nil
}

func (k *keychain) resolveDirectory() (string, error) {
	if k.directory == "" {
		return "", fmt.Errorf("no directory provided for file keychain")
	}

	stat, err := os.Stat(k.directory)
	if os.IsNotExist(err) {
		err = os.MkdirAll(k.directory, 0700)
	} else if err != nil {
		return "", err
	} else if !stat.IsDir() {
		return "", fmt.Errorf("%s is a file, not a directory", k.directory)
	}

	return k.directory, err
}

func (k *keychain) unlock() error {
	if k.password != "" || k.passwordFunc == nil {
		return nil
	}

	password, err := k.passwordFunc(fmt.Sprintf("Enter passphrase to unlock %q", k.directory))
	if err != nil {
		return err
	}

	k.password = password
	return nil
}
