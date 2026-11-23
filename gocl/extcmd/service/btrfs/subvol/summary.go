package subvol

import (
	"bufio"
	"cmp"
	"errors"
	"fmt"
	"io"
	"iter"
	"maps"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

const btrfsSubvolListParseRegexpPattern = `(ID\s+(?P<id>[0-9]+))|` +
	`(gen\s+(?P<gen>[0-9]+))|` +
	`(cgen\s+(?P<cgen>[0-9]+))|` +
	`(parent\s+(?P<parent>[0-9]+))|` +
	`(top level\s+(?P<top_level>[0-9]+))|` +
	`(parent_uuid\s+(?P<snapshot_parent>[0-9A-Fa-f-]{1,36}))|` +
	`(received_uuid\s+(?P<rcv_uuid>[0-9A-Fa-f-]{1,36}))|` +
	`(uuid\s+(?P<uuid>[0-9A-Fa-f-]{36}))|` +
	`(path\s+(?P<path>[\pL\pN\pZ\pP\pS]+))`

var btrfsSubvolListParseRegexp = regexp.MustCompile(btrfsSubvolListParseRegexpPattern)

type Summary struct {
	isReadOnly *bool
	name       string

	Path      string
	ReachPath string
	snapshots []Summary

	Generation         int64
	ID                 int64
	CreationGeneration int64
	ParentID           int64
	TopLevelID         int64

	ParentUUID   uuid.UUID
	ReceivedUUID uuid.UUID
	UUID         uuid.UUID

	hasSnapshotsDataEnriched bool
}

func (v Summary) Name() string {
	if v.name == "" {
		return filepath.Base(v.Path)
	}

	return v.name
}

func (v Summary) WithPath(val string) Summary {
	v.Path = val

	return v
}

func (v Summary) ResetSnapshots() Summary {
	v.hasSnapshotsDataEnriched = false
	v.snapshots = nil

	return v
}

func (v Summary) ResetReachPath() Summary {
	v.ReachPath = ""

	return v
}

func (v Summary) ResetUUID() Summary {
	v.UUID = uuid.Nil

	return v
}

func (v Summary) IsCreatedAsSnapshot() bool {
	return v.ParentUUID != uuid.Nil || v.ReceivedUUID != uuid.Nil
}

func (v Summary) IDString() string { return strconv.FormatInt(v.ID, 10) }

func (v Summary) ShortString() string {
	sb := &strings.Builder{}

	v.writeShortString(sb)

	return sb.String()
}

func (v Summary) String() string {
	sb := strings.Builder{}

	sb.WriteByte('(')

	sb.WriteString("ID: ")
	sb.WriteString(strconv.FormatInt(v.ID, 10))

	sb.WriteString(", gen: ")
	sb.WriteString(strconv.FormatInt(v.Generation, 10))

	sb.WriteString(", cgen: ")
	sb.WriteString(strconv.FormatInt(v.CreationGeneration, 10))

	sb.WriteString(", UUID: ")
	sb.WriteString(v.UUID.String())

	if len(v.ReachPath) != 0 {
		sb.WriteString(", reach path: ")
		sb.WriteString(v.ReachPath)
	} else {
		sb.WriteString(", path: ")
		sb.WriteString(placeholder.UnspecifiedIfEmpty(v.Path))
	}

	if v.hasSnapshotsDataEnriched {
		sb.WriteString(", snapshots count: ")
		sb.WriteString(strconv.FormatInt(int64(len(v.snapshots)), 10))
	}

	sb.WriteByte(')')

	return sb.String()
}

// Snapshots returns sequence of snapshots of this subvolume.
// Consider, sequence sorted in such order that most recent snapshot comes first (see [Summary.CreationGeneration]).
func (v Summary) Snapshots() (iter.Seq2[int, Summary], error) {
	if v.hasSnapshotsDataEnriched {
		return slices.All(v.snapshots), nil
	}

	return nil, fmt.Errorf("subvolume '%s' has not yet enriched with snapshots data", v)
}

func (v Summary) writeShortString(sb *strings.Builder) {
	sb.WriteByte('(')

	sb.WriteString("ID: ")
	sb.WriteString(strconv.FormatInt(v.ID, 10))

	if v.ReachPath != "" {
		sb.WriteString(", reach path: ")
		sb.WriteString(v.ReachPath)
	} else if v.Path != "" {
		sb.WriteString(", path: ")
		sb.WriteString(v.Path)
	}

	if len(v.snapshots) != 0 {
		sb.WriteString(", snapshots count: ")
		sb.WriteString(strconv.FormatInt(int64(len(v.snapshots)), 10))
	}

	sb.WriteByte(')')
}

func (v Summary) CalcSending(received SummaryList) (res SubvolSend, err error) {
	return v.calcSending(received.ReceivedUUIDLookup())
}

func (v Summary) calcSending(receivedLookup SummaryReceivedUUIDLookup) (res SubvolSend, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to calculate subvolume send list relatively specified received subvolumes (snapshots); subvolume: %s; %w", v, err)
		}
	}()

	subvol := v

	if !subvol.hasSnapshotsDataEnriched {
		return SubvolSend{}, errors.New("subvolume has not yet enriched with snapshots data")
	}

	recentReceivedSnapshotIdx := -1

	for i, snapshot := range subvol.snapshots {
		if found, ok := receivedLookup[snapshot.UUID]; ok {
			if len(found) != 1 {
				return SubvolSend{}, fmt.Errorf("unexpected count '%d' of received items found by received UUID '%s', one item expected", len(found), snapshot.UUID)
			}

			recentReceivedSnapshotIdx = i

			break
		}
	}

	if recentReceivedSnapshotIdx == -1 {
		recentReceivedSnapshotIdx = len(subvol.snapshots)
	}

	var sendSeq SendSequence

	for i := recentReceivedSnapshotIdx - 1; i > -1; i-- {
		item := SendItem{Snapshot: subvol.snapshots[i]}

		if i+1 < len(v.snapshots) {
			item.ParentSnapshot = &subvol.snapshots[i+1]
		}

		sendSeq = append(sendSeq, item)
	}

	return SubvolSend{Subvol: v, Sequence: sendSeq}, nil
}

// nolint:nestif,gocognit,maintidx
func (v *Summary) FromBtrfsSubvolumeShowOutputReader(r io.Reader) (err error) {
	if v == nil {
		return nil
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to map '%s' from 'btrfs subvolume show' command output; %w", typestring.TypeStringForExpr(v), err)
		}
	}()

	const (
		cGenerationField  = "Gen at creation:"
		flagsField        = "Flags:"
		generationField   = "Generation:"
		nameField         = "Name:"
		parentIDField     = "Parent ID:"
		parentUUIDField   = "Parent UUID:"
		pathField         = "<path>"
		receivedUUIDField = "Received UUID:"
		subvolumeIDField  = "Subvolume ID:"
		topLevelIDField   = "Top level ID:"
		uuidField         = "UUID:"
	)

	var (
		s                = bufio.NewScanner(r)
		cp               = Summary{}
		unobservedFields = map[string]struct{}{
			cGenerationField:  {},
			generationField:   {},
			nameField:         {},
			parentIDField:     {},
			parentUUIDField:   {},
			pathField:         {},
			receivedUUIDField: {},
			subvolumeIDField:  {},
			topLevelIDField:   {},
			uuidField:         {},
		}
	)

	lineCounter := 0

	for s.Scan() {
		lineCounter++

		switch lineCounter {
		case 1:
			// subvolume path
			cp.Path = string(s.Bytes())

			delete(unobservedFields, pathField)

			continue
		}

		if len(s.Bytes()) > 512 {
			return fmt.Errorf("%w; line too long, line '%d'", errs.ErrInvalidSyntax, lineCounter)
		}

		line := strings.TrimSpace(s.Text())
		parseErr := error(nil)

		if valStr, ok := strings.CutPrefix(line, uuidField); ok {
			if cp.UUID, parseErr = uuid.Parse(strings.TrimSpace(valStr)); parseErr != nil {
				return fmt.Errorf("%w; unable to parse field '%s' as UUID (line '%s'); %w", errs.ErrInvalidSyntax, uuidField, line, parseErr)
			} else if cp.UUID == uuid.Nil {
				return fmt.Errorf("%w; unable to parse field '%s' as UUID (line '%s'); parsed UUID can't be '%s'", errs.ErrInvalidSyntax, uuidField, line, uuid.Nil)
			}

			delete(unobservedFields, uuidField)
		} else if valStr, ok := strings.CutPrefix(line, parentUUIDField); ok {
			switch valStr = strings.TrimSpace(valStr); valStr == "-" {
			case true:
			default:
				if cp.ParentUUID, parseErr = uuid.Parse(valStr); parseErr != nil {
					return fmt.Errorf("%w; unable to parse field '%s' as UUID (line '%s'); %w", errs.ErrInvalidSyntax, parentUUIDField, line, parseErr)
				} else if cp.ParentUUID == uuid.Nil {
					return fmt.Errorf("%w; unable to parse field '%s' as UUID (line '%s'); parsed UUID can't be '%s'", errs.ErrInvalidSyntax, parentUUIDField, line, uuid.Nil)
				}
			}

			delete(unobservedFields, parentUUIDField)
		} else if valStr, ok := strings.CutPrefix(line, receivedUUIDField); ok {
			switch valStr = strings.TrimSpace(valStr); valStr == "-" {
			case true:
			default:
				if cp.ReceivedUUID, parseErr = uuid.Parse(valStr); parseErr != nil {
					return fmt.Errorf("%w; unable to parse field '%s' as UUID (line '%s'); %w", errs.ErrInvalidSyntax, receivedUUIDField, line, parseErr)
				} else if cp.ReceivedUUID == uuid.Nil {
					return fmt.Errorf("%w; unable to parse field '%s' as UUID (line '%s'); parsed UUID can't be '%s'", errs.ErrInvalidSyntax, receivedUUIDField, line, uuid.Nil)
				}
			}

			delete(unobservedFields, receivedUUIDField)
		} else if valStr, ok := strings.CutPrefix(line, subvolumeIDField); ok {
			if cp.ID, parseErr = strconv.ParseInt(strings.TrimSpace(valStr), 10, 64); parseErr != nil {
				return fmt.Errorf("%w; unable to parse field '%s' as int64 (line '%s'); %w", errs.ErrInvalidSyntax, subvolumeIDField, line, parseErr)
			}

			delete(unobservedFields, subvolumeIDField)
		} else if valStr, ok := strings.CutPrefix(line, generationField); ok {
			if cp.Generation, parseErr = strconv.ParseInt(strings.TrimSpace(valStr), 10, 64); parseErr != nil {
				return fmt.Errorf("%w; unable to parse field '%s' as int64 (line '%s'); %w", errs.ErrInvalidSyntax, generationField, line, parseErr)
			} else if cp.Generation < 1 {
				return fmt.Errorf("%w; unable to parse field '%s' as int64 (line '%s'); parsed value can't be less than 1", errs.ErrInvalidSyntax, generationField, line)
			}

			delete(unobservedFields, generationField)
		} else if valStr, ok := strings.CutPrefix(line, cGenerationField); ok {
			if cp.CreationGeneration, parseErr = strconv.ParseInt(strings.TrimSpace(valStr), 10, 64); parseErr != nil {
				return fmt.Errorf("%w; unable to parse field '%s' as int64 (line '%s'); %w", errs.ErrInvalidSyntax, cGenerationField, line, parseErr)
			} else if cp.CreationGeneration < 1 {
				return fmt.Errorf("%w; unable to parse field '%s' as int64 (line '%s'); parsed value can't be less than 1", errs.ErrInvalidSyntax, cGenerationField, line)
			}

			delete(unobservedFields, cGenerationField)
		} else if valStr, ok := strings.CutPrefix(line, parentIDField); ok {
			if cp.ParentID, parseErr = strconv.ParseInt(strings.TrimSpace(valStr), 10, 64); parseErr != nil {
				return fmt.Errorf("%w; unable to parse field '%s' as int64 (line '%s'); %w", errs.ErrInvalidSyntax, parentIDField, line, parseErr)
			}

			delete(unobservedFields, parentIDField)
		} else if valStr, ok := strings.CutPrefix(line, topLevelIDField); ok {
			if cp.TopLevelID, parseErr = strconv.ParseInt(strings.TrimSpace(valStr), 10, 64); parseErr != nil {
				return fmt.Errorf("%w; unable to parse field '%s' as int64 (line '%s'); %w", errs.ErrInvalidSyntax, topLevelIDField, line, parseErr)
			}

			delete(unobservedFields, topLevelIDField)
		} else if valStr, ok := strings.CutPrefix(line, nameField); ok {
			if cp.name = strings.TrimSpace(valStr); cp.name == "" {
				return fmt.Errorf("%w; unable to parse field '%s' (line '%s'); name can't be an empty string", errs.ErrInvalidSyntax, nameField, line)
			}

			delete(unobservedFields, nameField)
		} else if valStr, ok := strings.CutPrefix(line, flagsField); ok {
			valStr = strings.TrimSpace(valStr)

			switch valStr {
			case "readonly":
				cp.isReadOnly = new(true)
			case "":
				cp.isReadOnly = new(false)
			default:
				return fmt.Errorf("%w; unable to extract read-only mark from field '%s' (line '%s')", errs.ErrInvalidSyntax, flagsField, line)
			}

			delete(unobservedFields, flagsField)
		}
	}

	if err = s.Err(); err != nil {
		return fmt.Errorf("error occurred while scanning reader; %w", err)
	}

	if len(unobservedFields) != 0 {
		return fmt.Errorf(
			"%w; some of required fields '%s' missed in the output",
			errs.ErrInvalidSyntax,
			strings.Join(slices.Collect(maps.Keys(unobservedFields)), "', '"))
	}

	if cp.Generation < cp.CreationGeneration {
		return fmt.Errorf(
			"%w; unexpected condition occurred, subvolume generation '%d' less than creation generation '%d'",
			errs.ErrInvalidSyntax,
			cp.Generation,
			cp.CreationGeneration)
	}

	*v = cp

	return nil
}

type SummaryList []Summary

func (v SummaryList) String() string {
	if v == nil {
		return placeholder.Nil
	} else if len(v) == 0 {
		return placeholder.Empty
	}

	sb := &strings.Builder{}

	sb.WriteByte('[')

	sb.WriteString("count: ")
	sb.WriteString(strconv.Itoa(len(v)))

	for i := range len(v) {
		sb.WriteString(", ")

		v[i].writeShortString(sb)
	}

	sb.WriteByte(']')

	return sb.String()
}

func (v SummaryList) Indices(idx []int) SummaryList {
	if idx == nil {
		return nil
	} else if len(idx) == 0 {
		switch v {
		case nil:
			return nil
		default:
			return SummaryList{}
		}
	}

	return func() SummaryList {
		defer func() {
			if panicErr := errs.PanicErr(recover()); panicErr != nil {
				panicErr = fmt.Errorf("re-panic, some of input indices is out of range; %w", panicErr)

				panic(panicErr)
			}
		}()

		res := make(SummaryList, 0, len(idx))

		for _, i := range idx {
			res = append(res, v[i])
		}

		return res
	}()
}

func (v SummaryList) First(fn func(Summary) bool) (Summary, bool) {
	if v == nil {
		return Summary{}, false
	}

	for i := range len(v) {
		if fn(v[i]) {
			return v[i], true
		}
	}

	return Summary{}, false
}

func (v SummaryList) All(fn func(Summary) bool) SummaryList {
	if v == nil {
		return nil
	}

	res := make(SummaryList, 0, len(v))

	for i := range len(v) {
		if fn(v[i]) {
			res = append(res, v[i])
		}
	}

	return res
}

func (v SummaryList) Count() int { return len(v) }

// CreatedAsSnapshot returns iterator over subvolumes which are created as snapshot operation.
func (v SummaryList) CreatedAsSnapshot() iter.Seq[Summary] {
	return func(yield func(Summary) bool) {
		for i := range len(v) {
			if !v[i].IsCreatedAsSnapshot() {
				continue
			}

			if !yield(v[i]) {
				return
			}
		}
	}
}

func (v SummaryList) SortByCreationGeneration(desc bool) {
	if desc {
		slices.SortFunc(v, func(a, b Summary) int { return cmp.Compare(b.CreationGeneration, a.CreationGeneration) })

		return
	}

	slices.SortFunc(v, func(a, b Summary) int { return cmp.Compare(a.CreationGeneration, b.CreationGeneration) })
}

//nolint:gocognit,maintidx
func (v *SummaryList) FromBtrfsSubvolumeListOutputReader(r io.Reader) (err error) {
	if v == nil {
		return
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to map '%s' from 'btrfs subvolume list' command output; %w", typestring.TypeStringForExpr(v), err)
		}
	}()

	s := bufio.NewScanner(r)
	cp := make(SummaryList, 0, 64) // nolint:mnd,revive

	type fieldParser struct {
		fn          func(in string, item *Summary) error
		name        string
		subexpIndex int
	}

	fieldsParsers := []fieldParser{
		{
			name:        "subvolume ID",
			subexpIndex: btrfsSubvolListParseRegexp.SubexpIndex("id"),
			fn: func(in string, item *Summary) error {
				val, parseErr := strconv.ParseInt(in, 10, 64)
				if parseErr != nil {
					return fmt.Errorf("%w; failed to parse '%s' as int64; %w", errs.ErrInvalidSyntax, in, parseErr)
				}

				item.ID = val

				return nil
			},
		},
		{
			name:        "generation",
			subexpIndex: btrfsSubvolListParseRegexp.SubexpIndex("gen"),
			fn: func(in string, item *Summary) error {
				val, parseErr := strconv.ParseInt(in, 10, 64)
				if parseErr != nil {
					return fmt.Errorf("%w; failed to parse '%s' as int64; %w", errs.ErrInvalidSyntax, in, parseErr)
				} else if val < 1 {
					return fmt.Errorf("%w; parsed value can't be less than 1", errs.ErrInvalidSyntax)
				}

				item.Generation = val

				return nil
			},
		},
		{
			name:        "creation generation",
			subexpIndex: btrfsSubvolListParseRegexp.SubexpIndex("cgen"),
			fn: func(in string, item *Summary) error {
				val, parseErr := strconv.ParseInt(in, 10, 64)
				if parseErr != nil {
					return fmt.Errorf("%w; failed to parse '%s' as int64; %w", errs.ErrInvalidSyntax, in, parseErr)
				} else if val < 1 {
					return fmt.Errorf("%w; parsed value can't be less than 1", errs.ErrInvalidSyntax)
				}

				item.CreationGeneration = val

				return nil
			},
		},
		{
			name:        "parent ID",
			subexpIndex: btrfsSubvolListParseRegexp.SubexpIndex("parent"),
			fn: func(in string, item *Summary) error {
				val, parseErr := strconv.ParseInt(in, 10, 64)
				if parseErr != nil {
					return fmt.Errorf("%w; failed to parse '%s' as int64; %w", errs.ErrInvalidSyntax, in, parseErr)
				}

				item.ParentID = val

				return nil
			},
		},
		{
			name:        "top level ID",
			subexpIndex: btrfsSubvolListParseRegexp.SubexpIndex("top_level"),
			fn: func(in string, item *Summary) error {
				val, parseErr := strconv.ParseInt(in, 10, 64)
				if parseErr != nil {
					return fmt.Errorf("%w; failed to parse '%s' as int64; %w", errs.ErrInvalidSyntax, in, parseErr)
				}

				item.TopLevelID = val

				return nil
			},
		},
		{
			name:        "snapshot parent UUID",
			subexpIndex: btrfsSubvolListParseRegexp.SubexpIndex("snapshot_parent"),
			fn: func(in string, item *Summary) error {
				if in == "-" {
					item.ParentUUID = uuid.Nil

					return nil
				}

				val, parseErr := uuid.Parse(in)
				if parseErr != nil {
					return fmt.Errorf("%w; failed to parse '%s' as UUID; %w", errs.ErrInvalidSyntax, in, parseErr)
				} else if val == uuid.Nil {
					return fmt.Errorf("%w; parsed UUID can't be '%s'", errs.ErrInvalidSyntax, uuid.Nil)
				}

				item.ParentUUID = val

				return nil
			},
		},
		{
			name:        "received UUID",
			subexpIndex: btrfsSubvolListParseRegexp.SubexpIndex("rcv_uuid"),
			fn: func(in string, item *Summary) error {
				if in == "-" {
					item.ReceivedUUID = uuid.Nil

					return nil
				}

				val, parseErr := uuid.Parse(in)
				if parseErr != nil {
					return fmt.Errorf("%w; failed to parse '%s' as UUID; %w", errs.ErrInvalidSyntax, in, parseErr)
				} else if val == uuid.Nil {
					return fmt.Errorf("%w; parsed UUID can't be '%s'", errs.ErrInvalidSyntax, uuid.Nil)
				}

				item.ReceivedUUID = val

				return nil
			},
		},
		{
			name:        "subvolume UUID",
			subexpIndex: btrfsSubvolListParseRegexp.SubexpIndex("uuid"),
			fn: func(in string, item *Summary) error {
				val, parseErr := uuid.Parse(in)
				if parseErr != nil {
					return fmt.Errorf("%w; failed to parse '%s' as UUID; %w", errs.ErrInvalidSyntax, in, parseErr)
				} else if val == uuid.Nil {
					return fmt.Errorf("%w; parsed UUID can't be '%s'", errs.ErrInvalidSyntax, uuid.Nil)
				}

				item.UUID = val

				return nil
			},
		},
		{
			name:        "subvolume path",
			subexpIndex: btrfsSubvolListParseRegexp.SubexpIndex("path"),
			fn: func(in string, item *Summary) error {
				const maxLen = 4096

				if len(in) > maxLen {
					return fmt.Errorf("%w; too long, max. allowed is '%d'", errs.ErrInvalidSyntax, maxLen)
				}

				item.Path = in

				return nil
			},
		},
	}

	lineParser := func(line []byte) (item Summary, localErr error) {
		m := btrfsSubvolListParseRegexp.FindAllStringSubmatch(string(line), -1)

		for i := range len(fieldsParsers) {
			found := false

			for j, v := range m {
				if v[fieldsParsers[i].subexpIndex] != "" {
					switch localErr = fieldsParsers[i].fn(v[fieldsParsers[i].subexpIndex], &item); localErr != nil {
					case true:
						return Summary{}, fmt.Errorf("error while parsing field '%s' in text '%s'; %w", fieldsParsers[i].name, line, localErr)
					}

					m = slices.Delete(m, j, j+1)

					found = true

					break
				}
			}

			if !found {
				return Summary{}, fmt.Errorf("no field '%s' found in text '%s'", fieldsParsers[i].name, line)
			}
		}

		if item.Generation < item.CreationGeneration {
			return Summary{},
				fmt.Errorf(
					"%w; unexpected condition occurred, subvolume generation '%d' less than creation generation '%d'",
					errs.ErrInvalidSyntax,
					item.Generation,
					item.CreationGeneration)
		}

		return item, nil
	}

	lineCounter := 0
	for s.Scan() {
		lineCounter++

		item, itemErr := lineParser(s.Bytes())
		if itemErr != nil {
			return fmt.Errorf("failed to parse line '%d'; %w", lineCounter, itemErr)
		}

		cp = append(cp, item)
	}

	if err = s.Err(); err != nil {
		return fmt.Errorf("error occurred while scanning reader; %w", err)
	}

	*v = cp

	return nil
}

func (v SummaryList) enrichSnapshotsData() {
	allCreatedAsSnapshot := slices.Collect(v.CreatedAsSnapshot())

	for i := range len(v) {
		subvol := v[i]

		subvol.snapshots = nil
		subvol.hasSnapshotsDataEnriched = true

		j := 0
		for j < len(allCreatedAsSnapshot) {
			if allCreatedAsSnapshot[j].ParentUUID == subvol.UUID {
				subvol.snapshots = append(subvol.snapshots, allCreatedAsSnapshot[j])

				allCreatedAsSnapshot = slices.Delete(allCreatedAsSnapshot, j, j+1)

				continue
			}

			j++
		}

		slices.SortFunc(subvol.snapshots, func(a, b Summary) int { return cmp.Compare(b.CreationGeneration, a.CreationGeneration) })

		v[i] = subvol
	}
}

func (v SummaryList) ReceivedUUIDLookup() SummaryReceivedUUIDLookup {
	if v == nil {
		return nil
	}

	res := make(SummaryReceivedUUIDLookup, len(v))

	for _, s := range v {
		res[s.ReceivedUUID] = append(res[s.ReceivedUUID], s)
	}

	return res
}

func (v SummaryList) SendList(received SummaryList) (SubvolSendList, error) {
	if v == nil {
		return nil, nil
	} else if len(v) == 0 {
		return SubvolSendList{}, nil
	}

	res := make(SubvolSendList, 0, len(v))
	receivedUUIDLookup := received.ReceivedUUIDLookup()

	for i := range len(v) {
		item, itemErr := v[i].calcSending(receivedUUIDLookup)
		if itemErr != nil {
			return nil, itemErr
		}

		res = append(res, item)
	}

	return res, nil
}

type SummaryReceivedUUIDLookup map[uuid.UUID]SummaryList
