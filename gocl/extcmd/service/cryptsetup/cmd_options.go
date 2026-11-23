package cryptsetup

import "rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"

type cmdOptionsBuilder struct{ dict stringsext.KVDictString }

func CmdOptions() cmdOptionsBuilder { return cmdOptionsBuilder{dict: stringsext.KVDictString{}} }

func (v cmdOptionsBuilder) Option(name, val string) cmdOptionsBuilder {
	if val == "" {
		delete(v.dict, name)

		return v
	}

	v.dict[name] = stringsext.ListString{val}

	return v
}

func (v cmdOptionsBuilder) Type(val string) cmdOptionsBuilder {
	if val == "" {
		delete(v.dict, "type")

		return v
	}

	v.dict["type"] = stringsext.ListString{val}

	return v
}

func (v cmdOptionsBuilder) KeyFile(val string) cmdOptionsBuilder {
	if val == "" {
		delete(v.dict, "key-file")

		return v
	}

	v.dict["key-file"] = stringsext.ListString{val}

	return v
}

func (v cmdOptionsBuilder) KeySlot(val *uint8) cmdOptionsBuilder {
	if val == nil {
		delete(v.dict, "key-slot")

		return v
	}

	v.dict["key-slot"] = stringsext.ListString{stringsext.FormatUintPtr(val)}

	return v
}

func (v cmdOptionsBuilder) Options() (res stringsext.KVOpts, err error) {
	err = res.FromKVDict(v.dict)

	return
}

func (v cmdOptionsBuilder) MustOptions() stringsext.KVOpts {
	res, err := v.Options()
	if err != nil {
		panic(err)
	}

	return res
}
