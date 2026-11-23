package query

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"rn.m.vo7.io/vo7io/m/gocl/misc/tagext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	db_comm "rn.m.vo7.io/vo7io/m/gocl/storage"
)

func ToPositionalSyntax(q string, paramNameQualifier ParamQualifier, syntax PositionalSyntaxVariant, args ...any) (posQ string, posArgs []any, err error) {
	if len(q) == 0 {
		return
	}

	posQ, posQBindings, posQParamPlaceCount, err := bindNamedParamsToPositions(q, paramNameQualifier, syntax)
	if err != nil {
		return
	}

	if syntax == QuestionSignPositionalSyntaxVariant && posQParamPlaceCount != len(posQBindings) {
		err = fmt.Errorf("specified positional syntax not allowed to use same parameter name miltiple times in query text; syntax: %s", syntax)

		return
	}

	if len(posQBindings) == 0 {
		// no bindings == no named parameters in source query.
		//
		return
	}

	argsAsNamedVals, err := tagext.CollectValuesSlice(args, db_comm.FieldDbTagName)
	if err != nil {
		err = fmt.Errorf("error while extracting named query params' values from specified args; err: %w", err)

		return
	}

	posArgs = make([]any, len(posQBindings))

	for paramName, paramPos := range posQBindings {
		if paramVal, ok := argsAsNamedVals[paramName]; ok {
			posArgs[paramPos] = paramVal

			continue
		}

		err = fmt.Errorf("can't bind query parameter to value from specified args; there is no field in args matching query parameter name; parameter name: %s; err: %w",
			paramName,
			errs.ErrNamedQueryParamMissing)

		return
	}

	return
}

//nolint:gocognit,maintidx,gocyclo,cyclop
func bindNamedParamsToPositions(
	q string,
	paramNameQualifier ParamQualifier,
	syntax PositionalSyntaxVariant,
) (posQ string, bindings map[string]int, posQParamPlaceCount int, err error) {
	paramNameQualifierRune := rune(paramNameQualifier)

	if paramNameQualifierRune == '_' ||
		paramNameQualifierRune == '-' ||
		unicode.IsDigit(paramNameQualifierRune) ||
		unicode.IsLetter(paramNameQualifierRune) {
		err = fmt.Errorf("invalid query parameter name qualifier '%s'", string(paramNameQualifierRune))

		return
	}

	switch syntax {
	case DollarSignPositionalSyntaxVariant, QuestionSignPositionalSyntaxVariant:
	case UnspecifiedPositionalSyntaxVariant:
		err = errors.New("positional syntax variant not specified")

		return
	default:
		err = errs.NotSupportedWrap(fmt.Errorf("specified positional syntax variant not supported; variant: %s", syntax))

		return
	}

	if len(q) == 0 {
		return
	}

	type scopeType int

	const (
		queryScope     scopeType = 1
		paramNameScope scopeType = 2

		paramNameBuildCap = 64
	)

	bindings = map[string]int{}
	posQBuilder := strings.Builder{}
	currNameBuilder := strings.Builder{}
	currScope := queryScope

	posQBuilder.Grow(len(q))
	currNameBuilder.Grow(paramNameBuildCap)

	onParamNameScopeTermination := func(termSymbol rune, termPos int) error {
		if currNameBuilder.Len() == 0 {
			return fmt.Errorf("empty query param name found; query: %s; position: %d; err: %w", q, termPos-1, errs.ErrInvalidQueryParamName)
		}

		currName := currNameBuilder.String()

		currNameBindPos, ok := bindings[currName]
		if !ok {
			currNameBindPos = len(bindings)
			bindings[currName] = currNameBindPos
		}

		switch syntax {
		case DollarSignPositionalSyntaxVariant:
			posQBuilder.WriteRune('$')
			posQBuilder.WriteString(strconv.Itoa(currNameBindPos + 1))
		case QuestionSignPositionalSyntaxVariant:
			posQBuilder.WriteRune('?')
		}

		posQParamPlaceCount++

		if termSymbol != -1 {
			posQBuilder.WriteRune(termSymbol)
		}

		currNameBuilder.Reset()
		currNameBuilder.Grow(paramNameBuildCap)

		currScope = queryScope

		return nil
	}

	for symPos, sym := range q {
		switch sym {
		case paramNameQualifierRune:
			if currScope == queryScope {
				currScope = paramNameScope

				continue
			}

			if currNameBuilder.Len() == 0 {
				currScope = queryScope

				posQBuilder.WriteRune(paramNameQualifierRune)

				continue
			}

			if err = onParamNameScopeTermination(-1, symPos); err != nil {
				return
			}

			currScope = paramNameScope
		case '-', '_':
			if currScope == queryScope {
				posQBuilder.WriteRune(sym)

				continue
			}

			if currNameBuilder.Len() == 0 && sym == '-' {
				err = fmt.Errorf("query param name can't start with '-' symbol; query: %s; position: %d; err: %w", q, symPos, errs.ErrInvalidQueryParamName)

				return
			}

			currNameBuilder.WriteRune(sym)
		default:
			if currScope == queryScope {
				posQBuilder.WriteRune(sym)

				continue
			}

			if unicode.IsDigit(sym) {
				if currNameBuilder.Len() == 0 {
					err = fmt.Errorf("query param name can't start with digit symbol; query: %s; position: %d; err: %w", q, symPos, errs.ErrInvalidQueryParamName)

					return
				}

				currNameBuilder.WriteRune(sym)

				continue
			}

			if unicode.IsLetter(sym) {
				currNameBuilder.WriteRune(sym)

				continue
			}

			if err = onParamNameScopeTermination(sym, symPos); err != nil {
				return
			}
		}
	}

	if currScope == paramNameScope {
		if err = onParamNameScopeTermination(-1, len(q)); err != nil {
			return
		}
	}

	posQ = posQBuilder.String()

	return
}
