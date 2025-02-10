package appstore

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/majd/ipatool/pkg/http"
	"github.com/majd/ipatool/pkg/util"
	"github.com/pkg/errors"
)

type LoginAddressResult struct {
	FirstName string `plist:"firstName,omitempty"`
	LastName  string `plist:"lastName,omitempty"`
}

type LoginAccountResult struct {
	Email   string             `plist:"appleId,omitempty"`
	Address LoginAddressResult `plist:"address,omitempty"`
}

type LoginResult struct {
	FailureType         string             `plist:"failureType,omitempty"`
	CustomerMessage     string             `plist:"customerMessage,omitempty"`
	Account             LoginAccountResult `plist:"accountInfo,omitempty"`
	DirectoryServicesID string             `plist:"dsPersonId,omitempty"`
	PasswordToken       string             `plist:"passwordToken,omitempty"`
}

func (a *appstore) Login(email, password, authCode string) error {
	if password == "" && !a.interactive {
		return ErrPasswordRequired
	}

	if password == "" && a.interactive {
		a.logger.Log().Msg("enter password:")

		var err error
		password, err = a.promptForPassword()
		if err != nil {
			return errors.Wrap(err, ErrGetData.Error())
		}
	}

	guid := util.MakeGuid(email)
	a.logger.Verbose().Str("guid", guid).Send()

	acc, err := a.login(email, password, authCode, guid, 0, false)
	if err != nil {
		return errors.Wrap(err, ErrLogin.Error())
	}

	a.logger.Log().
		Str("name", acc.Name).
		Str("email", acc.Email).
		Bool("success", true).
		Send()

	return nil
}

func (a *appstore) login(email, password, authCode, guid string, attempt int, failOnAuthCodeRequirement bool) (Account, error) {
	a.logger.Verbose().
		Int("attempt", attempt).
		Str("password", password).
		Str("email", email).
		Str("authCode", util.IfEmpty(authCode, "<nil>")).
		Msg("sending login request")

	redirect := ""
	var err error
	retry := true
	var res http.Result[LoginResult]

	for attempt := 1; retry && attempt <= 4; attempt++ {
		request := a.loginRequest(email, password, authCode, guid, attempt)
		request.URL, redirect = util.IfEmpty(redirect, request.URL), ""
		res, err = a.loginClient.Send(request)
		if err != nil {
			return Account{}, fmt.Errorf("request failed: %w", err)
		}

		if retry, redirect, err = a.parseLoginResponse(&res, attempt, authCode); err != nil {
			if errors.Is(err, ErrAuthCodeRequired) {
				if failOnAuthCodeRequirement {
					return Account{}, ErrAuthCodeRequired
				}

				if a.interactive {
					a.logger.Log().Msg("enter 2FA code:")
					authCode, err = a.promptForAuthCode()
					if err != nil {
						return Account{}, errors.Wrap(err, ErrGetData.Error())
					}

					return a.login(email, password, authCode, guid, 0, failOnAuthCodeRequirement)
				} else {
					a.logger.Log().Msg("2FA code is required; run the command again and supply a code using the `--auth-code` flag")
					return Account{}, nil
				}
			} else {
				return Account{}, err
			}
		}
	}

	if retry {
		return Account{}, errors.New("too many attempts")
	}

	sf, err := res.GetHeader(HTTPHeaderStoreFront)
	if err != nil {
		return Account{}, fmt.Errorf("failed to get storefront header: %w", err)
	}

	addr := res.Data.Account.Address
	acc := Account{
		Name:                strings.Join([]string{addr.FirstName, addr.LastName}, " "),
		Email:               res.Data.Account.Email,
		PasswordToken:       res.Data.PasswordToken,
		DirectoryServicesID: res.Data.DirectoryServicesID,
		StoreFront:          sf,
		Password:            password,
	}

	data, err := json.Marshal(acc)
	if err != nil {
		return Account{}, fmt.Errorf("failed to marshal json: %w", err)
	}

	err = a.keychain.Set("account", data)
	if err != nil {
		return Account{}, fmt.Errorf("failed to save account in keychain: %w", err)
	}

	return acc, nil
}

func (a *appstore) parseLoginResponse(res *http.Result[LoginResult], attempt int, authCode string) (retry bool, redirect string, err error) {
	if res.StatusCode == 302 {
		if redirect, err = res.GetHeader("location"); err != nil {
			err = fmt.Errorf("failed to retrieve redirect location: %w", err)
		} else {
			retry = true
		}
	} else if attempt == 1 && res.Data.FailureType == FailureTypeInvalidCredentials {
		retry = true
	} else if res.Data.FailureType == "" && authCode == "" && res.Data.CustomerMessage == CustomerMessageBadLogin {
		err = ErrAuthCodeRequired
	} else if res.Data.FailureType != "" {
		if res.Data.CustomerMessage != "" {
			err = errors.New(res.Data.CustomerMessage)
		} else {
			err = errors.New(res.Data.CustomerMessage)
		}
	} else if res.StatusCode != 200 {
		err = fmt.Errorf("got non-200 status code: %d", res.StatusCode)
	} else if res.Data.PasswordToken == "" {
		err = errors.New("PasswordToken is empty")
	} else if res.Data.DirectoryServicesID == "" {
		err = errors.New("DirectoryServicesID is empty")
	}
	return
}

func (a *appstore) loginRequest(email, password, authCode, guid string, attempt int) http.Request {
	return http.Request{
		Method:         http.MethodPOST,
		URL:            fmt.Sprintf("https://%s%s", PrivateAppStoreAPIDomain, PrivateAppStoreAPIPathAuthenticate),
		ResponseFormat: http.ResponseFormatXML,
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Payload: &http.XMLPayload{
			Content: map[string]interface{}{
				"appleId":       email,
				"attempt":       strconv.Itoa(attempt),
				"createSession": "true",
				"guid":          guid,
				"password":      fmt.Sprintf("%s%s", password, authCode),
				"rmp":           "0",
				"why":           "signIn",
			},
		},
	}
}

func (a *appstore) promptForAuthCode() (string, error) {
	reader := bufio.NewReader(a.ioReader)
	authCode, err := reader.ReadString('\n')
	if err != nil {
		return "", errors.Wrap(err, ErrGetData.Error())
	}

	authCode = strings.Trim(authCode, "\n")
	authCode = strings.Trim(authCode, "\r")

	return authCode, nil
}

func (a *appstore) promptForPassword() (string, error) {
	password, err := a.machine.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", errors.Wrap(err, ErrGetData.Error())
	}

	return string(password), nil
}
