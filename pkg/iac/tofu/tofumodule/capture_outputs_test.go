package tofumodule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnwrapTofuOutputs_Values(t *testing.T) {
	doc := []byte(`{
		"host": {"value": "10.0.0.5", "type": "string", "sensitive": false},
		"password": {"value": "s3cr3t", "type": "string", "sensitive": true},
		"ports": {"value": [6379, 6380], "type": ["list", "number"], "sensitive": false},
		"labels": {"value": {"tier": "cache"}, "type": ["object", {"tier": "string"}], "sensitive": false}
	}`)

	values, err := unwrapTofuOutputs(doc)
	assert.NoError(t, err)

	assert.Equal(t, "10.0.0.5", values["host"])
	assert.Equal(t, "s3cr3t", values["password"])
	assert.Len(t, values["ports"], 2)
	assert.Equal(t, map[string]interface{}{"tier": "cache"}, values["labels"])
}

func TestUnwrapTofuOutputs_EmptyDocumentIsNotAnError(t *testing.T) {
	for _, doc := range []string{"", "{}", "  \n"} {
		values, err := unwrapTofuOutputs([]byte(doc))
		assert.NoError(t, err, "doc %q", doc)
		assert.Empty(t, values)
	}
}

func TestUnwrapTofuOutputs_NonEnvelopeShapeErrors(t *testing.T) {
	_, err := unwrapTofuOutputs([]byte(`{"host": "bare-value-not-an-envelope"}`))
	assert.Error(t, err)
}
