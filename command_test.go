package enumgen

import (
	"strings"
	"testing"
)

func TestUsageTabs(t *testing.T) {
	if strings.Contains(Usage, "\t") {
		t.Fail()
	}
}
