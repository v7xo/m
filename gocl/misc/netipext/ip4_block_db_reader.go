package netipext

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type IP4BlockDbReader interface {
	Format() IP4BlockDbFormat
	ReadAll(context.Context, io.Reader) (IP4BlockRecordList, error)
}

type ip2LocationDb1CSV struct{}

var _ IP4BlockDbReader = &ip2LocationDb1CSV{}

func (*ip2LocationDb1CSV) Format() IP4BlockDbFormat { return IP4BlockDbFormatLiteIP2Location1CSV }

func (*ip2LocationDb1CSV) ReadAll(ctx context.Context, r io.Reader) (res IP4BlockRecordList, err error) {
	log := logext.NopLogger

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to read IP4 block database, format '%s'; %w", IP4BlockDbFormatLiteIP2Location1CSV, err)

			log.Err(ctx, err).Send()

			return
		}

		log.Info(ctx).Msgf("succeeded to read IP4 block database, format '%s'; totally '%d' records", IP4BlockDbFormatLiteIP2Location1CSV, len(res))
	}()

	ctx = attrname.DictContext(ctx, attrname.Dict{attrname.LogScope: "lite-ip2location-1-csv"})

	csvReader := csv.NewReader(r)
	csvReader.FieldsPerRecord = 4 // nolint
	csvReader.ReuseRecord = true
	csvReader.TrimLeadingSpace = true

	line := 0
	res = make(IP4BlockRecordList, 0, 1024) // nolint

	if log, err = contextext.Log(ctx); err != nil {
		return nil, err
	}

	for {
		select {
		case <-ctx.Done():
			return nil, context.Cause(ctx)
		default:
		}

		line++

		rec, readErr := csvReader.Read()
		if errors.Is(readErr, io.EOF) {
			break
		} else if readErr != nil {
			return nil, readErr
		}

		start, parseErr := strconv.ParseUint(rec[0], 10, 32)
		if parseErr != nil {
			return nil, fmt.Errorf("%w; invalid range start, at line '%d', field '1'; %w", errs.ErrInvalidSyntax, line, parseErr)
		}

		end, parseErr := strconv.ParseUint(rec[1], 10, 32)
		if parseErr != nil {
			return nil, fmt.Errorf("%w; invalid range end, at line '%d', field '2'; %w", errs.ErrInvalidSyntax, line, parseErr)
		}

		if start > end {
			return nil, fmt.Errorf("%w; invalid range '[%d; %d]', at line '%d', field '2'", errs.ErrInvalidSyntax, start, end, line)
		}

		if rec[2] == "" {
			return nil, fmt.Errorf("%w; empty marker, at line '%d', field '3'", errs.ErrInvalidSyntax, line)
		}

		res = append(
			res,
			IP4BlockRecord{
				Block: IP4Block{
					start: uint32(start),
					end:   uint32(end),
				},
				Tag: rec[2],
			},
		)

		if line%1000 == 0 {
			log.Trace(ctx).Msgf("succeeded to read first '%d' records", line)
		}
	}

	return res, nil
}

func NewIP4BlockDbReader(format IP4BlockDbFormat) (_ IP4BlockDbReader, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of IP4 block database reader, format '%s'; %w", format, err)
		}
	}()

	switch format {
	case IP4BlockDbFormatLiteIP2Location1CSV:
		return &ip2LocationDb1CSV{}, nil
	case IP4BlockDbFormatUnspecified:
		fallthrough
	default:
		return nil, fmt.Errorf("unspecified/unknown format '%s'", format)
	}
}
