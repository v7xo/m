package startconf

const (
	ConfFlenvSelfFlagPrefix = "st"

	DefaultTestDataDirName = "test-data"

	DevelopConfEntryDir          = `.`
	DevelopConfYamlFileName      = `conf-develop.yml`
	DevelopLocalConfYamlFileName = `conf-develop-local.yml`

	TestConfEntryDir          = "./" + DefaultTestDataDirName
	TestConfYamlFileName      = `conf-test.yml`
	TestLocalConfYamlFileName = `conf-test-local.yml`
)
