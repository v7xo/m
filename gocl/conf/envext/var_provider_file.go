package envext

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// nolint:tagliatelle
type varProviderFileModel struct {
	Vars map[string]*string `yaml:"environment"`
	Root bool               `yaml:"root"`
}

type varProviderFile struct {
	entryDir string
	names    []string
}

var _ VarProvider = &varProviderFile{}

func (s *varProviderFile) VariablesDict(ctx context.Context) (_ map[string]string, err error) {
	entryDir, names := s.entryDir, s.names

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to load configuration files starting from '%s' directory; %w", entryDir, err)
		}
	}()

	groups, err := s.buildFileGroups(ctx, entryDir, names...)
	if err != nil {
		return
	}

	res := map[string]string{}

	for i := range len(groups) {
		select {
		case <-ctx.Done():
			return nil, context.Cause(ctx)
		default:
		}

		group := groups[i]

		for _, path := range group {
			select {
			case <-ctx.Done():
				return nil, context.Cause(ctx)
			default:
			}

			var (
				dataYaml []byte
				data     *varProviderFileModel
			)

			dataYaml, err = os.ReadFile(path)
			if err != nil {
				return
			}

			yamlDec := yaml.NewDecoder(bytes.NewBuffer(dataYaml))
			yamlDec.KnownFields(true)

			if err = yamlDec.Decode(&data); err != nil {
				return nil, fmt.Errorf("failed to unmarshal content of file '%s' as YAML; %w", path, err)
			}

			for varName, varVal := range data.Vars {
				if _, ok := res[varName]; ok || varVal == nil {
					continue
				}

				res[varName] = *varVal
			}

			if data.Root {
				break
			}
		}
	}

	return res, nil
}

func (*varProviderFile) buildFileGroups(ctx context.Context, entryDirPath string, names ...string) ([][]string, error) {
	if len(names) == 0 {
		return nil, nil
	}

	entryDirPath, err := filepath.Abs(entryDirPath)
	if err != nil {
		return nil, err
	}

	var entryDir fs.FileInfo

	// search first existing dir upward specified entryDirPath.
	for currentDirPath := entryDirPath; ; {
		select {
		case <-ctx.Done():
			return nil, context.Cause(ctx)
		default:
		}

		var err error

		entryDir, err = os.Stat(currentDirPath)
		if os.IsNotExist(err) {
			if parentDir := filepath.Dir(currentDirPath); parentDir != currentDirPath {
				currentDirPath = parentDir

				continue
			}

			return nil, err
		} else if err != nil {
			return nil, err
		}

		break
	}

	if !entryDir.IsDir() {
		return nil, fmt.Errorf("specified path '%s' doesn't point to valid directory", entryDirPath)
	}

	res := make([][]string, 0, len(names))

	for i := range names {
		select {
		case <-ctx.Done():
			return nil, context.Cause(ctx)
		default:
		}

		if len(names[i]) == 0 {
			continue
		}

		if filepath.IsAbs(names[i]) {
			res = append(res, []string{names[i]})

			continue
		}

		var (
			nameGroup      []string
			currentDirPath = entryDirPath
		)

		for {
			select {
			case <-ctx.Done():
				return nil, context.Cause(ctx)
			default:
			}

			path := filepath.Join(currentDirPath, names[i])

			stat, err := os.Stat(path)
			if os.IsNotExist(err) {
				err = nil //nolint
			} else if err != nil {
				return nil, err
			} else if !stat.IsDir() {
				nameGroup = append(nameGroup, path)
			}

			if parentDir := filepath.Dir(currentDirPath); parentDir != currentDirPath {
				currentDirPath = parentDir

				continue
			}

			break
		}

		res = append(res, nameGroup)
	}

	return res, nil
}

func VarProviderFile(entryDir string, names ...string) VarProvider {
	return &varProviderFile{entryDir: entryDir, names: names}
}
