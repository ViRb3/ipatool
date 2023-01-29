package cmd

import (
	"github.com/99designs/keyring"
	"github.com/juju/persistent-cookiejar"
	"github.com/majd/ipatool/pkg/appstore"
	"github.com/majd/ipatool/pkg/keychain"
	"github.com/majd/ipatool/pkg/log"
	"github.com/majd/ipatool/pkg/util"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
)

func newCookieJar() (*cookiejar.Jar, error) {
	machine := util.NewMachine(util.MachineArgs{
		OperatingSystem: util.NewOperatingSystem(),
	})
	jar, err := cookiejar.New(&cookiejar.Options{
		Filename: filepath.Join(machine.HomeDirectory(), ConfigDirectoryName, CookieJarFileName),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cookie jar")
	}

	return jar, nil
}

func keyringBackendType() keyring.BackendType {
	return keyring.FileBackend
}

func newKeyring(logger log.Logger, passphrase string, interactive bool) (keyring.Keyring, error) {
	machine := util.NewMachine(util.MachineArgs{
		OperatingSystem: util.NewOperatingSystem(),
	})

	ring, err := keyring.Open(keyring.Config{
		AllowedBackends: []keyring.BackendType{
			keyringBackendType(),
		},
		ServiceName: KeychainServiceName,
		FileDir:     filepath.Join(machine.HomeDirectory(), ConfigDirectoryName),
		FilePasswordFunc: func(s string) (string, error) {
			return "", nil
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to open keyring")
	}

	return ring, nil
}

func configureConfigDirectory() error {
	os := util.NewOperatingSystem()
	machine := util.NewMachine(util.MachineArgs{
		OperatingSystem: os,
	})

	configDirectoryPath := filepath.Join(machine.HomeDirectory(), ConfigDirectoryName)
	_, err := os.Stat(configDirectoryPath)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(configDirectoryPath, 0700)
		if err != nil {
			return errors.Wrap(err, "failed to create config directory")
		}
	} else if err != nil {
		return errors.Wrap(err, "could not read metadata")
	}

	return nil
}

func parseOutputFormat(value string) (OutputFormat, error) {
	switch value {
	case "json":
		return OutputFormatJSON, nil
	case "text":
		return OutputFormatText, nil
	default:
		return OutputFormatJSON, errors.Errorf("invalid output format: %s", value)
	}
}

func newLogger(format OutputFormat, verbose bool) log.Logger {
	var writer io.Writer

	switch format {
	case OutputFormatJSON:
		writer = zerolog.SyncWriter(os.Stdout)
	case OutputFormatText:
		writer = log.NewWriter()
	}

	return log.NewLogger(log.LoggerArgs{
		Verbose: verbose,
		Writer:  writer,
	})
}

func newAppStore(
	cmd *cobra.Command,
	keychainPassphrase string,
) (appstore.AppStore, error) {
	logger := cmd.Context().Value("logger").(log.Logger)
	interactive := cmd.Context().Value("interactive").(bool)

	cookieJar, err := newCookieJar()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cookie jar")
	}

	os := util.NewOperatingSystem()

	keyring, err := newKeyring(logger, keychainPassphrase, interactive)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create keyring")
	}

	return appstore.NewAppStore(appstore.AppStoreArgs{
		Logger:    logger,
		CookieJar: cookieJar,
		Keychain: keychain.NewKeychain(keychain.KeychainArgs{
			Keyring: keyring,
		}),
		Interactive: interactive,
		Machine: util.NewMachine(util.MachineArgs{
			OperatingSystem: os,
		}),
		OperatingSystem: os,
	}), nil
}

func parseID(bundleID string, appID int64) (any, error) {
	if (bundleID == "" && appID == -1) || (bundleID != "" && appID != -1) {
		return nil, errors.New("required exactly one of flags \"bundle-identifier\", \"app-id\"")
	}
	if bundleID != "" {
		return bundleID, nil
	}
	return appID, nil
}
