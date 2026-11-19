package docs

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpecIsValidJSON(t *testing.T) {
	var v map[string]interface{}
	assert.NoError(t, json.Unmarshal(Spec, &v))
	assert.Contains(t, v, "paths")
}
