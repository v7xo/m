package contract

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/huandu/xstrings"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/value/valonce"
)

type Describer interface {
	Describe(ctx context.Context, typ reflect.Type) (Contract, error)
	IsLocalNameAttr(reflect.Type) bool
	IsNameAttr(reflect.Type) bool
	IsNamespaceAttr(reflect.Type) bool
	IsNamespaceLocator(reflect.Type) (bool, LocatorKind)
}

type describer struct {
	fopt.LogOpt

	cache *sync.Map
}

func (*describer) IsNameAttr(typ reflect.Type) bool {
	return typ != nil &&
		typ.Kind() == reflect.Struct &&
		typ.Implements(reflect.TypeFor[nameAttr]())
}

func (*describer) IsNamespaceAttr(typ reflect.Type) bool {
	return typ != nil &&
		typ.Kind() == reflect.Struct &&
		typ.Implements(reflect.TypeFor[namespaceAttr]())
}

func (*describer) IsLocalNameAttr(typ reflect.Type) bool {
	return typ != nil &&
		typ.Kind() == reflect.Struct &&
		typ.Implements(reflect.TypeFor[localNameAttr]())
}

func (*describer) IsNamespaceLocator(typ reflect.Type) (_ bool, kind LocatorKind) {
	if typ == nil || typ.Kind() != reflect.Struct || !typ.Implements(reflect.TypeFor[NamespaceLocator]()) {
		return false, LocatorKindUnspecified
	}

	switch {
	case typ == reflect.TypeFor[AsPackage]():
		return true, LocatorKindAsPackage
	case typ == reflect.TypeFor[AsTag]():
		return true, LocatorKindAsTag
	case typ == reflect.TypeFor[AsType]():
		return true, LocatorKindAsType
	default:
		return true, LocatorKindRedirect
	}
}

func (*describer) IsLocalNameLocator(typ reflect.Type) (_ bool, kind LocatorKind) {
	if typ == nil || typ.Kind() != reflect.Struct || !typ.Implements(reflect.TypeFor[LocalNameLocator]()) {
		return false, LocatorKindUnspecified
	}

	switch {
	case typ == reflect.TypeFor[AsType]():
		return true, LocatorKindAsType
	case typ == reflect.TypeFor[AsTag]():
		return true, LocatorKindAsTag
	default:
		return true, LocatorKindRedirect
	}
}

func (s *describer) Describe(ctx context.Context, typ reflect.Type) (Contract, error) {
	valAny, ok := s.cache.Load(typ)
	if !ok {
		valAny, _ = s.cache.LoadOrStore(typ, s.describeOnce(typ))
	}

	return valAny.(valonce.Of[Contract]).Eval(ctx) //nolint:forcetypeassert
}

func (s *describer) describeOnce(typ reflect.Type) valonce.Of[Contract] {
	return valonce.New(
		func(ctx context.Context) (c Contract, err error) {
			defer func() {
				if err != nil {
					err = fmt.Errorf("%w; failed to describe contract for type '%s'; %w", ErrContract, stringsext.TypeString(typ), err)

					s.Log().Err(ctx, err).Send()

					return
				}

				s.Log().Trace(ctx).Msgf("succeeded to describe poly-contract for type '%s'; contract: %s", stringsext.TypeString(typ), c)
			}()

			if err = s.validateTypeForDescribe(typ); err != nil {
				return
			}

			nameAttrField, err := s.nameAttrField(typ)
			if err != nil {
				return
			}

			c = Contract{Type: typ, NameAttr: nameAttrField}

			if c.NamespaceLocator, err = s.resolveLocator(c.Type, c.NameAttr, false); err != nil {
				return
			}

			if c.Namespace, err = s.calcLocatorValue(c.NamespaceLocator, false); err != nil {
				return
			}

			if c.LocalNameLocator, err = s.resolveLocator(c.Type, c.NameAttr, true); err != nil {
				return
			}

			if c.LocalName, err = s.calcLocatorValue(c.LocalNameLocator, true); err != nil {
				return
			}

			c.Name = c.Namespace + c.LocalName

			return c, nil
		})
}

func (*describer) calcLocatorValue(loc *AttrLocator, localName bool) (_ string, err error) {
	choiceMsg := "namespace"

	if localName {
		choiceMsg = "local name"
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to calculate value of %s from locator '%s'; %w", choiceMsg, loc, err)
		}
	}()

	val := loc.Raw
	if val == "" {
		switch {
		case loc.AttrTag != nil:
			val = loc.AttrTag.Value
		}
	} else if loc.AttrTag != nil && loc.AttrTag.Value != "" && loc.AttrTag.Value != val {
		return "", fmt.Errorf("conflict occurred, locator itself supplies a value '%s' and its tag does same '%s'", loc.Raw, loc.AttrTag.Value)
	}

	if val == "" {
		val = loc.Origin.Name()
	}

	if (loc.Kind.IsAsPackage() || (loc.AttrKind.IsNamespace() && loc.Kind.IsAsType())) &&
		!strings.HasSuffix(val, "/") {
		val += "/"
	}

	if loc.AttrTag != nil {
		//nolint:exhaustive
		switch loc.AttrTag.Flags & TagFlagCaseMask {
		case TagFlagCamelCase:
			val = xstrings.ToCamelCase(val)
		case TagFlagKebabCase:
			val = xstrings.ToKebabCase(val)
		case TagFlagSnakeCase:
			val = xstrings.ToSnakeCase(val)
		case TagFlagNone:
		default:
			return "", fmt.Errorf("case transformation '%s' is unknown", loc.AttrTag.Flags&TagFlagCaseMask)
		}
	}

	return val, nil
}

func (s *describer) resolveLocator(typ reflect.Type, nameAttr reflect.StructField, localName bool) (link *AttrLocator, err error) {
	attrKind := AttrKindNamespace
	if localName {
		attrKind = AttrKindLocalName
	}

	currLink := &AttrLocator{Origin: typ, AttrField: nameAttr, AttrKind: attrKind}

	for {
		locatorField, locatorKind, locatorErr := s.locatorField(currLink.AttrField.Type, localName)
		if locatorErr != nil {
			return
		}

		currLink.Kind = locatorKind

		if err = s.populateAttrFieldTags(typ, currLink, locatorKind.IsAsTag()); err != nil {
			return
		}

		switch locatorKind {
		case LocatorKindAsPackage:
			currLink.Raw = currLink.Origin.PkgPath()
		case LocatorKindAsType:
			currLink.Raw = currLink.Origin.Name()
		case LocatorKindAsTag:
			currLink.Raw = currLink.AttrTag.Value
		case LocatorKindRedirect:
			var nextAttrField reflect.StructField

			if localName {
				nextAttrField, err = s.localNameAttrField(locatorField.Type)
			} else {
				nextAttrField, err = s.namespaceAttrField(locatorField.Type)
			}

			if err != nil {
				return
			}

			nextLink := AttrLocator{
				AttrField: nextAttrField,
				AttrKind:  currLink.AttrKind,
				Backward:  currLink,
				Origin:    locatorField.Type,
			}
			currLink.Forward = &nextLink
			currLink = &nextLink

			continue
		case LocatorKindUnspecified:
			fallthrough
		default:
			return nil, fmt.Errorf("locator kind '%s' unspecified/unsupported", locatorKind)
		}

		if err = s.validateAttrLocator(currLink); err != nil {
			return
		}

		link = currLink

		return
	}
}

func (s *describer) locatorField(typ reflect.Type, localName bool) (_ reflect.StructField, _ LocatorKind, err error) {
	var (
		choiceMsg         = "namespace"
		choiceFieldName   = locatorNamespaceFieldName
		choiceLocatorTest = s.IsNamespaceLocator
	)

	if localName {
		choiceMsg = "local name"
		choiceFieldName = locatorLocalNameFieldName
		choiceLocatorTest = s.IsLocalNameLocator
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to locate the field designating contract %s locator in type '%s'; %w", choiceMsg, stringsext.TypeString(typ), err)
		}
	}()

	var (
		locatorField *reflect.StructField
		locatorKind  LocatorKind
	)

	for fi := range min(locatorFieldSearchLimit, typ.NumField()) {
		currField := typ.Field(fi)
		if currField.Name != choiceFieldName {
			continue
		}

		isLocator := false

		if isLocator, locatorKind = choiceLocatorTest(currField.Type); !isLocator {
			continue
		}

		locatorField = &currField

		break
	}

	if locatorField == nil {
		return reflect.StructField{}, LocatorKindUnspecified, fmt.Errorf("%w; no field designating contract %s", errs.ErrNotFound, choiceMsg)
	}

	return *locatorField, locatorKind, nil
}

func (s *describer) localNameAttrField(typ reflect.Type) (_ reflect.StructField, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to locate the field designating contract local name in type '%s'; %w", stringsext.TypeString(typ), err)
		}
	}()

	var (
		foundField    *reflect.StructField
		foundFieldIdx = -1
	)

	for i := range typ.NumField() {
		curr := typ.Field(i)

		if !s.IsLocalNameAttr(curr.Type) {
			continue
		}

		if foundField != nil {
			return reflect.StructField{},
				fmt.Errorf("%w; there are two fields which designate contract local name at least: '%s' (at '%d' position) and '%s' (at '%d' position)",
					errs.ErrAmbiguousMatch,
					curr.Name,
					i,
					foundField.Name,
					foundFieldIdx)
		}

		foundField, foundFieldIdx = &curr, i
	}

	if foundField == nil {
		return reflect.StructField{}, fmt.Errorf("%w; no field designating contract local name", errs.ErrNotFound)
	}

	return *foundField, nil
}

func (s *describer) namespaceAttrField(typ reflect.Type) (_ reflect.StructField, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to locate the field designating contract namespace in type '%s'; %w", stringsext.TypeString(typ), err)
		}
	}()

	var (
		foundField    *reflect.StructField
		foundFieldIdx = -1
	)

	for i := range typ.NumField() {
		curr := typ.Field(i)

		if !s.IsNamespaceAttr(curr.Type) {
			continue
		}

		if foundField != nil {
			return reflect.StructField{},
				fmt.Errorf("%w; there are two fields which designate contract namespace at least: '%s' (at '%d' position) and '%s' (at '%d' position)",
					errs.ErrAmbiguousMatch,
					curr.Name,
					i,
					foundField.Name,
					foundFieldIdx)
		}

		foundField, foundFieldIdx = &curr, i
	}

	if foundField == nil {
		return reflect.StructField{}, fmt.Errorf("%w; no field designating contract namespace", errs.ErrNotFound)
	}

	return *foundField, nil
}

func (s *describer) nameAttrField(typ reflect.Type) (_ reflect.StructField, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to locate the field designating contract name in type '%s'; %w", stringsext.TypeString(typ), err)
		}
	}()

	var (
		foundField    *reflect.StructField
		foundFieldIdx = -1
	)

	for i := range typ.NumField() {
		curr := typ.Field(i)

		if !s.IsNameAttr(curr.Type) {
			continue
		}

		if foundField != nil {
			return reflect.StructField{},
				fmt.Errorf("%w; there are two fields which designate contract name at least: '%s' (at '%d' position) and '%s' (at '%d' position)",
					errs.ErrAmbiguousMatch,
					curr.Name,
					i,
					foundField.Name,
					foundFieldIdx)
		}

		foundField, foundFieldIdx = &curr, i
	}

	if foundField == nil {
		return reflect.StructField{}, fmt.Errorf("%w; no field designating contract name", errs.ErrNotFound)
	}

	return *foundField, nil
}

func (s *describer) populateAttrFieldTags(typ reflect.Type, loc *AttrLocator, required bool) (err error) {
	if loc.LocalNameTag, err = s.parseAttrFieldTag(typ, loc.AttrField, TagKeyLocalName); err != nil {
		return err
	}

	if loc.NamespaceTag, err = s.parseAttrFieldTag(typ, loc.AttrField, TagKeyNamespace); err != nil {
		return err
	}

	requredTagKey := placeholder.Unspecified

	if loc.AttrKind.IsLocalName() {
		loc.AttrTag = loc.LocalNameTag
		requredTagKey = TagKeyLocalName
	} else if loc.AttrKind.IsNamespace() {
		loc.AttrTag = loc.NamespaceTag
		requredTagKey = TagKeyNamespace
	}

	if required && loc.AttrTag == nil {
		fldMsg := ""
		if loc.AttrField.Name != "" && loc.AttrField.Name != "_" {
			fldMsg = " on field '" + loc.AttrField.Name + "'"
		}

		return fmt.Errorf("tag%s not contains key '%s'", fldMsg, requredTagKey)
	}

	return nil
}

func (s *describer) parseAttrFieldTag(typ reflect.Type, fld reflect.StructField, tagKey string) (_ *Tag, err error) {
	fldMsg := ""
	if fld.Name != "" && fld.Name != "_" {
		fldMsg = " on field '" + fld.Name + "'"
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to parse tag key '%s'%s of type '%s'; %w", tagKey, fldMsg, stringsext.TypeString(typ), err)
		}
	}()

	tagRaw, ok := fld.Tag.Lookup(tagKey)
	if !ok {
		return
	}

	var (
		dict        stringsext.KVDictString
		dictKeysPos []string
	)

	if dict, dictKeysPos, err = stringsext.ParseKVDictString(tagRaw); err != nil {
		return nil, fmt.Errorf("%w; failed to parse value '%s'; %w", errs.ErrFieldTagInvalidSyntax, tagRaw, err)
	}

	var (
		tagValue string
		tagFlags TagFlag
	)

	for pos := range len(dictKeysPos) {
		dictKey := dictKeysPos[pos]

		if pos == 0 {
			if len(dict[dictKey]) != 0 {
				return nil,
					fmt.Errorf("%w; first option '%s' always considered as value itself, but it declares a value(s) '%s'",
						errs.ErrFieldTagInvalidSyntax,
						dictKey,
						dict[dictKey])
			}

			tagValue = dictKey

			continue
		}

		if dictKey == "" {
			continue
		}

		tagFlag, tagFlagErr := ParseTagFlag(dictKey)
		if tagFlagErr == nil {
			if len(dict[dictKey]) != 0 {
				return nil,
					fmt.Errorf("%w; option '%s' always considered as a flag, but it declares a value(s) '%s'",
						errs.ErrFieldTagInvalidSyntax,
						dictKey,
						dict[dictKey])
			}

			tagFlags |= tagFlag

			continue
		}

		// it might be an option with value, but not now (only flags supprted)
		return nil, fmt.Errorf("%w; unknown option '%s' at '%d' position", errs.ErrFieldTagInvalidSyntax, dictKey, pos)
	}

	tagValueTrimmed := strings.TrimSpace(tagValue)
	if tagValue != "" && tagValueTrimmed == "" {
		return nil, fmt.Errorf("%w; value contains white-space symbols only", errs.ErrFieldTagInvalidSyntax)
	}

	tagValue = tagValueTrimmed

	if err = tagFlags.Validate(); err != nil {
		return nil, fmt.Errorf("%w; %w", errs.ErrFieldTagInvalidSyntax, err)
	}

	return &Tag{
		Raw:   tagRaw,
		Value: tagValue,
		Flags: tagFlags,
	}, nil
}

func (*describer) validateTypeForDescribe(typ reflect.Type) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("invalid type '%s'; %w", stringsext.TypeString(typ), err)
			err = errs.JoinIfNot(err, errs.ErrBadArgs)
		}
	}()

	if typ == nil {
		return fmt.Errorf("%w; type argument is nil", errs.ErrNilValue)
	}

	if typ.Kind() != reflect.Struct {
		return errors.New("not a struct type")
	}

	return
}

func (*describer) validateAttrLocator(loc *AttrLocator) (err error) {
	choiceMsg := ""

	switch loc.AttrKind {
	case AttrKindLocalName:
		choiceMsg = "local name"
	case AttrKindNamespace:
		choiceMsg = "namespace"
	case AttrKindName, AttrKindUnspecified:
		fallthrough
	default:
		return fmt.Errorf("unable to validate attribute locator, invalid attribute kind '%s'", loc.AttrKind)
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("contract %s locator not valid; %w", choiceMsg, err)
		}
	}()

	curr := loc
	for curr != nil {
		if curr.AttrKind != loc.AttrKind {
			return fmt.Errorf("invalid attribute kind '%s', '%s' expected; locator: %s", curr.AttrKind, loc.AttrKind, curr)
		}

		invalidTagKeyAnnotation := ""

		switch curr.Kind {
		case LocatorKindAsPackage, LocatorKindRedirect:
			if curr.Backward == nil {
				switch {
				case curr.AttrTag != nil:
					if curr.AttrKind.IsLocalName() {
						invalidTagKeyAnnotation = TagKeyLocalName

						break
					}

					invalidTagKeyAnnotation = TagKeyNamespace
				}

				break
			}

			switch {
			case curr.LocalNameTag != nil:
				invalidTagKeyAnnotation = TagKeyLocalName
			case curr.NamespaceTag != nil:
				invalidTagKeyAnnotation = TagKeyNamespace
			}
		case LocatorKindAsTag, LocatorKindAsType:
			break
		case LocatorKindUnspecified:
			fallthrough
		default:
			return fmt.Errorf("invalid locator kind '%s'; locator: %s", curr.Kind, curr)
		}

		if invalidTagKeyAnnotation != "" {
			fldNameMsg := ""

			if curr.AttrField.Name != "" && curr.AttrField.Name != "_" {
				fldNameMsg = " '" + curr.AttrField.Name + "'"
			}

			return fmt.Errorf("%w; field%s designating contract locator in type '%s' can't be annotated with tag key '%s'",
				errs.ErrFieldTagAnnotationNotAllowed,
				fldNameMsg,
				stringsext.TypeString(curr.Origin),
				invalidTagKeyAnnotation)
		}

		curr = curr.Backward
	}

	return nil
}

func NewDescriber(opts ...DescriberOptionFn) (_ Describer, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", stringsext.TypeStringFor[describer](), err)
			err = errs.JoinIfNot(err, ErrContract)
		}
	}()

	impl := &describer{cache: &sync.Map{}}
	if err = fopt.Apply(impl, opts...); err != nil {
		return nil, err
	}

	if err = impl.SetLogScope(logScope); err != nil {
		return nil, err
	}

	return impl, nil
}

type DescriberOptionFn func(*describer) error

func (o DescriberOptionFn) WithLog(val logext.Logger) DescriberOptionFn {
	return fopt.WithLog(o, val)
}

func DescriberOption() DescriberOptionFn { return fopt.Nop[describer] }
