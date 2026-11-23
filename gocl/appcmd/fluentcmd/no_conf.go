package fluentcmd

import "rn.m.vo7.io/vo7io/m/gocl/validation"

type noConf struct{}

func (noConf) Validate(validation.Validator) error { return nil }
