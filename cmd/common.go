package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	cookiejar "github.com/juju/persistent-cookiejar"
	"github.com/majd/ipatool/v2/pkg/appstore"
	"github.com/majd/ipatool/v2/pkg/http"
	"github.com/majd/ipatool/v2/pkg/keychain"
	"github.com/majd/ipatool/v2/pkg/log"
	"github.com/majd/ipatool/v2/pkg/util"
	"github.com/majd/ipatool/v2/pkg/util/machine"
	"github.com/majd/ipatool/v2/pkg/util/operatingsystem"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var dependencies = Dependencies{}
var keychainPassphrase string

type Dependencies struct {
	Logger    log.Logger
	OS        operatingsystem.OperatingSystem
	Machine   machine.Machine
	CookieJar http.CookieJar
	Keychain  keychain.Keychain
	AppStore  appstore.AppStore
}

// newLogger returns a new logger instance.
func newLogger(format OutputFormat, verbose bool) log.Logger {
	var writer io.Writer

	switch format {
	case OutputFormatJSON:
		writer = zerolog.SyncWriter(os.Stdout)
	case OutputFormatText:
		writer = log.NewWriter()
	}

	return log.NewLogger(log.Args{
		Verbose: verbose,
		Writer:  writer,
	},
	)
}

// newCookieJar returns a new cookie jar instance.
func newCookieJar(machine machine.Machine) http.CookieJar {
	cookiePath := filepath.Join(machine.HomeDirectory(), ConfigDirectoryName, CookieJarFileName)
	// Remove stale zero-byte lock files. The juju/go4/lock library only does
	// PID-based stale detection when the file has content; a 0-byte lock file
	// (left by a process killed between O_TRUNC and PID write) blocks forever.
	lockPath := cookiePath + ".lock"
	if fi, err := os.Stat(lockPath); err == nil && fi.Size() == 0 {
		os.Remove(lockPath)
	}
	return util.Must(cookiejar.New(&cookiejar.Options{
		Filename: cookiePath,
	}))
}

// newKeychain returns a new keychain instance.
func newKeychain(machine machine.Machine, logger log.Logger, interactive bool) keychain.Keychain {
	return keychain.New(keychain.Args{
		Directory: filepath.Join(machine.HomeDirectory(), ConfigDirectoryName),
	})
}

// initWithCommand initializes the dependencies of the command.
func initWithCommand(cmd *cobra.Command) {
	verbose := cmd.Flag("verbose").Value.String() == "true"
	interactive, _ := cmd.Context().Value(interactiveKey).(bool)
	format := util.Must(OutputFormatFromString(cmd.Flag("format").Value.String()))

	dependencies.Logger = newLogger(format, verbose)
	dependencies.OS = operatingsystem.New()
	dependencies.Machine = machine.New(machine.Args{OS: dependencies.OS})
	dependencies.CookieJar = newCookieJar(dependencies.Machine)
	dependencies.Keychain = newKeychain(dependencies.Machine, dependencies.Logger, interactive)
	dependencies.AppStore = appstore.NewAppStore(appstore.Args{
		CookieJar:       dependencies.CookieJar,
		OperatingSystem: dependencies.OS,
		Keychain:        dependencies.Keychain,
		Machine:         dependencies.Machine,
	})

	util.Must("", createConfigDirectory(dependencies.OS, dependencies.Machine))
}

// createConfigDirectory creates the configuration directory for the CLI tool, if needed.
func createConfigDirectory(os operatingsystem.OperatingSystem, machine machine.Machine) error {
	configDirectoryPath := filepath.Join(machine.HomeDirectory(), ConfigDirectoryName)
	_, err := os.Stat(configDirectoryPath)

	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(configDirectoryPath, 0700)
		if err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("could not read metadata: %w", err)
	}

	return nil
}
