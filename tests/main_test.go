// +build postgres

package tenancytests

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	defer closeDB()
	os.Exit(m.Run())
}
