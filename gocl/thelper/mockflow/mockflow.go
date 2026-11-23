package mockflow

import (
	"os"
	"strconv"
)

const MockFlowEnvVarName = "MOCK_FLOW"

func MockFlow() bool {
	val, ok := os.LookupEnv(MockFlowEnvVarName)
	if !ok {
		return false
	}

	valBool, err := strconv.ParseBool(val)
	if err != nil {
		return false
	}

	return valBool
}
