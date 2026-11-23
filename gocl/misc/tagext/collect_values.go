package tagext

import (
	"fmt"
	"reflect"
	"strconv"

	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

func CollectValuesSlice(vals []any, tag string) (res map[string]any, err error) {
	res = map[string]any{}

	for i := range len(vals) {
		if err = collectValues(vals[i], tag, ".["+strconv.Itoa(i)+"]", false, res); err != nil {
			return
		}
	}

	return
}

func CollectValues(val any, tag string) (res map[string]any, err error) {
	res = map[string]any{}

	if err = collectValues(val, tag, "", false, res); err != nil {
		return
	}

	return
}

func collectValues(val any, tag string, parentFld string, skipValNilErr bool, res map[string]any) (err error) {
	defer func() {
		if err != nil && parentFld != "" {
			err = fmt.Errorf("parent field '%s'; %w", parentFld, err)
		}
	}()

	valType := reflect.TypeOf(val)
	if valType == nil {
		if !skipValNilErr {
			return fmt.Errorf("%w; unable to get tag field for nil", errs.ErrNilValue)
		}

		return
	}

	for valType.Kind() == reflect.Pointer {
		valType = valType.Elem()
	}

	valValue := reflect.ValueOf(val)
	for valValue.Kind() == reflect.Pointer || valValue.Kind() == reflect.Interface {
		valValue = valValue.Elem()
	}

	if valType.Kind() != reflect.Struct {
		return
	}

	for fldIdx := 0; fldIdx < valType.NumField(); fldIdx++ {
		var (
			fld       reflect.StructField
			fldTagVal string
			ok        bool
		)

		fld = valType.Field(fldIdx)

		if !fld.IsExported() {
			continue
		} else if fldTagVal, ok = fld.Tag.Lookup(tag); !ok {
			if !valValue.IsValid() {
				continue
			}

			if err = collectValues(valValue.Field(fldIdx).Interface(), tag, parentFld+"."+fld.Name, true, res); err != nil {
				return
			}
		} else if fldTagVal == "-" {
			continue
		} else if fldTagVal == "" {
			return fmt.Errorf("%w; type '%s', field '%s', tag '%s'", errs.ErrFieldTagValueEmpty, valType, fld.Name, tag)
		} else if _, ok = res[fldTagVal]; ok {
			return fmt.Errorf("%w; type '%s', field '%s', tag '%s', tag value '%s'", errs.ErrFieldTagValueAmbiguous, valType, fld.Name, tag, fldTagVal)
		} else if valValue.IsValid() {
			res[fldTagVal] = valValue.Field(fldIdx).Interface()
		} else {
			res[fldTagVal] = nil
		}
	}

	return
}
