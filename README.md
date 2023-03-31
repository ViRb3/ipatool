# IPATool

`ipatool` is a command line tool that allows you to search for iOS apps on the [App Store](https://apps.apple.com) and download a copy of the app package, known as an _ipa_ file.

![Demo](./demo.gif)

This is a fork of the upstream project, adding a bunch of new features that improve usability and aid automation. For specifics, please check the commits or the usage below.

## Usage

### General

```bash
A cli tool for interacting with Apple's ipa files

Usage:
  ipatool [command]

Available Commands:
  auth        Authenticate with the App Store
  completion  Generate the autocompletion script for the specified shell
  download    Download (encrypted) iOS app packages from the App Store
  help        Help about any command
  lookup      Lookup information about a specific iOS app on the App Store
  purchase    Obtain a license for the app from the App Store
  search      Search for iOS apps available on the App Store

Flags:
      --format format     sets output format for command; can be 'text', 'json' (default text)
  -h, --help              help for ipatool
      --non-interactive   run in non-interactive session
      --verbose           enables verbose logs
  -v, --version           version for ipatool

Use "ipatool [command] --help" for more information about a command.
```

### App format

You can select apps by either app id (integer) or bundle id (string). For unlisted (removed) apps, if you have ever downloaded them before, you will be able to download them again only through the app id.

```bash
Flags:
  -i, --app-id int                   App ID of the target iOS app
  -b, --bundle-identifier string     The bundle identifier of the target iOS app
```

## License

IPATool is released under the [MIT license](https://github.com/majd/ipatool/blob/main/LICENSE).
