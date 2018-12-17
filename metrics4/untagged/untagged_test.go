package untagged

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseLabelsValues(t *testing.T) {
	labelsNames := []string{ "a", "b" }
	labels := []string { "a", "1", "b", "2" }
	labelsValues, err := parseLabelsValues(labelsNames, labels)
	if err != nil {
		t.Fatal(err)
	}
	assert.EqualValues(t, []string{ "1", "2" }, labelsValues)
}

func TestMakeNameFromLabels(t *testing.T) {
	labelsNames := []string{ "a", "b" }
	labels := []string { "a", "1", "b", "2" }
	name := makeNameFromLabels(labelsNames, labels)
	assert.EqualValues(t, "a_1_b_2", name)
}
