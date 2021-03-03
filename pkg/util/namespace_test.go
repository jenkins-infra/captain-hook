package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamespace(t *testing.T) {
	os.Setenv("POD_NAMESPACE", "foobar")
	ns, err := Namespace()
	assert.NoError(t, err)
	assert.Equal(t, "foobar", ns)
}
