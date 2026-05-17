package nntpReaderWriter

import (
	mock "github.com/Tensai75/nntpReaderWriter/testutils"
)

type TestStep struct {
	ScriptStep       mock.ScriptStep
	WriteData        []byte
	ExpectedReadData []byte
	ExpectedLines    []string
	ExpectedCode     int
	ExpectedMsg      string
	ExpectedError    error
}

type TestScript struct {
	Name  string
	Steps []TestStep
}

func NewTestReaderWriterWithTestScript(script TestScript, option NntpReaderWriterOptions) *NntpReaderWriter {
	var scriptSteps []mock.ScriptStep
	for _, step := range script.Steps {
		scriptSteps = append(scriptSteps, step.ScriptStep)
	}
	conn := mock.NewScriptedConn(scriptSteps)
	return NewNntpReaderWriter(conn, option)
}
