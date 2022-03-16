package tests

import (
	"github.com/mirogindev/gomer/test_uttils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFindObjectsToBuild(t *testing.T) {
	builder := test_uttils.CreateTestSchema()
	builder.SetDefaultScalars()
	inputs, outputs := builder.FindObjectsToBuild()
	assert.Equal(t, len(inputs), 13)
	assert.Equal(t, len(outputs), 6)
}

func TestBuildObjects(t *testing.T) {
	builder := test_uttils.CreateTestSchema()
	builder.SetDefaultScalars()
	builder.FindObjectsToBuild()
	builtInputs, builtOutputs := builder.CreateObjects()
	assert.Equal(t, len(builtInputs), 13)
	assert.Equal(t, len(builtOutputs), 6)
}

func TestBuildObjectsWithFields(t *testing.T) {
	builder := test_uttils.CreateTestSchema()
	builder.SetDefaultScalars()
	builder.FindObjectsToBuild()
	builder.CreateObjects()
	builtInputsWithFields, builtOutputsWithFields := builder.CreateObjectsFields()
	assert.Equal(t, len(builtInputsWithFields), 13)
	assert.Equal(t, len(builtOutputsWithFields), 6)
}

func TestBuildBuildSchema(t *testing.T) {
	builder := test_uttils.CreateTestSchema()
	_, err := builder.Build()
	assert.Empty(t, err)

}
