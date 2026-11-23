package value

import "strconv"

type JSONFloat64 float64

//nolint:mnd,revive
func (v JSONFloat64) MarshalJSON() ([]byte, error) {
	buf := []byte(strconv.FormatFloat(float64(v), 'f', -1, 64))

	for k := len(buf) - 1; k > -1; k-- {
		if buf[k] == '.' {
			return buf, nil
		}
	}

	l := len(buf)

	if cap(buf)-l > 1 {
		buf = buf[:l+2]
		buf[l] = '.'
		buf[l+1] = '0'

		return buf, nil
	}

	bufNew := make([]byte, l+2)
	copy(bufNew, buf)
	bufNew[l] = '.'
	bufNew[l+1] = '0'

	return bufNew, nil
}

//nolint:mnd,revive
func (v *JSONFloat64) UnmarshalJSON(b []byte) error {
	p, err := strconv.ParseFloat(string(b), 64)
	if err != nil {
		return err
	}

	*v = JSONFloat64(p)

	return nil
}
