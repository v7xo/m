package query

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/huandu/xstrings"

	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/value/valonce"
)

// Named represents special object declaring some query, associated name with it and other options.
type Named struct {
	evalTemplate        valonce.Of[*template.Template]
	name                string
	text                string
	paramQualifier      ParamQualifier
	hasPositionalSyntax bool
	evaluatable         bool
	evalOutputBufCap    int

	timeout timeout.Timeout
}

// Empty returns whether this instance is empty: no name, no text, etc.
func (nq Named) Empty() bool                    { return nq.name == "" }
func (nq Named) Anon() bool                     { return nq.name == anonNamedName }
func (nq Named) Evaluatable() bool              { return nq.evaluatable }
func (nq Named) HasPositionalSyntax() bool      { return nq.hasPositionalSyntax }
func (nq Named) Name() string                   { return nq.name }
func (nq Named) ParamQualifier() ParamQualifier { return nq.paramQualifier }
func (nq Named) Timeout() timeout.Timeout       { return nq.timeout }

func (nq Named) Text() (string, error) {
	if !nq.evaluatable {
		return nq.text, nil
	}

	return "", fmt.Errorf("storage query '%s' defined as evaluatable, use Eval method", nq.name)
}

func (nq Named) Eval(ctx context.Context, data any) (res Named, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to evaluate text of storage query '%s'; err: %w", nq.name, err)
		}
	}()

	if !nq.evaluatable {
		err = errors.New("query is not evaluatable")

		return
	}

	template, err := nq.evalTemplate.Eval(ctx)
	if err != nil {
		return
	}

	outBuf := bytes.NewBuffer(make([]byte, 0, nq.evalOutputBufCap))

	if err = template.Execute(outBuf, data); err != nil {
		err = fmt.Errorf("failed to execute Go text template; err: %w", err)

		return
	}

	res = nq
	res.evalTemplate = nil
	res.evaluatable = false
	res.text = outBuf.String()

	return
}

func (nq Named) MustEval(ctx context.Context, data any) Named {
	res, err := nq.Eval(ctx, data)
	if err != nil {
		panic(err)
	}

	return res
}

func (nq Named) String() string {
	if nq.evaluatable {
		return nq.name + ", Go template:\n" + nq.text
	}

	return nq.name + ":\n" + nq.text
}

func (nq Named) WithParamNameQualifier(v ParamQualifier) Named {
	nq.paramQualifier = v

	return nq
}

func (nq Named) WithPositionalSyntax(v bool) Named {
	nq.hasPositionalSyntax = v

	return nq
}

func (nq Named) WithTimeout(v timeout.Timeout) Named {
	nq.timeout = v

	return nq
}

func (nq *Named) prepareEvalTemplate() (err error) {
	if nq == nil || !nq.evaluatable {
		return
	}

	if nq.text == "" {
		err = errs.BadArgsWrapWith(errors.New("no query text specified"))

		return
	}

	goTemplateText := nq.text

	nq.evalTemplate = valonce.New(
		func(localCtx context.Context) (res *template.Template, localErr error) {
			defer func() {
				if localErr != nil {
					localErr = fmt.Errorf("failed to prepare Go template; err: %w", localErr)
				}
			}()

			txtFuncMap := sprig.TxtFuncMap()
			if _, ok := txtFuncMap["pascalcase"]; !ok {
				txtFuncMap["pascalcase"] = xstrings.ToPascalCase
			}

			t := template.New("").Funcs(txtFuncMap)

			res, localErr = t.Parse(goTemplateText)
			if localErr != nil {
				localErr = fmt.Errorf("failed to parse query text as Go text template; err: %w", localErr)
			}

			return
		},
	)

	return
}

func AnonNamedExt(paramNameQualifier ParamQualifier, positionalParams bool, text string) Named {
	return Named{name: anonNamedName, text: text, paramQualifier: paramNameQualifier, hasPositionalSyntax: positionalParams}
}

func AnonNamed(text string) Named {
	return Named{name: anonNamedName, text: text, paramQualifier: DefaultParamQualifier, hasPositionalSyntax: false}
}

func NewNamed(name string, text string, opts ...OptionFn) (res Named, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error while constructing named storage query; name: %s; err: %w", name, err)
		}
	}()

	if name == "" {
		err = errs.BadArgsWrapWith(errors.New("name of query can't be an empty string"))

		return
	}

	res = Named{
		name:             name,
		text:             text,
		paramQualifier:   DefaultParamQualifier,
		evalOutputBufCap: 4096, // nolint
	}

	if err = fopt.Apply(&res, opts...); err != nil {
		return
	}

	if res.evalOutputBufCap < 0 {
		err = errs.BadArgsWrapWith(fmt.Errorf("invalid value of evaluation output buffer, can't be negative; value: %d", res.evalOutputBufCap))

		return
	}

	err = res.prepareEvalTemplate()

	return
}

func MustNewNamed(name string, text string, opts ...OptionFn) Named {
	res, err := NewNamed(name, text, opts...)
	if err != nil {
		panic(err)
	}

	return res
}
