package hashauth

import (
	"encoding/hex"
	"errors"

	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type OptionFn func(*impl) error

func (o OptionFn) WithConfig(val Config) OptionFn {
	return func(i *impl) error {
		if err := o(i); err != nil {
			return err
		}

		salt, err := hex.DecodeString(val.SaltHex)
		if err != nil {
			return errors.New("failed to decode specified hex string of salt")
		}

		hash, err := hex.DecodeString(val.SecretHashHex)
		if err != nil {
			return errors.New("failed to decode specified hex string of secret's hash")
		}

		i.hash = hash
		i.salt = salt

		return nil
	}
}

func Option() OptionFn { return fopt.Nop[impl] }
