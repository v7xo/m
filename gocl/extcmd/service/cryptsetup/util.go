package cryptsetup

import clext_modl "rn.m.vo7.io/vo7io/m/gocl/extcmd/model"

var defaultCipherParams = clext_modl.MustDecodeCryptCipherParams("aes-xts-plain64")

func DefaultCipherParams() clext_modl.CryptCipherParams { return defaultCipherParams }
