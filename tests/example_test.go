package tenancytests

import (
	"context"

	"github.com/common-fate/tenancy"
)

func ExampleTTx() {
	ctx := context.TODO()
	conn := tenancy.Conn{}
	tx, _ := conn.BeginTx(ctx, nil)

	// an example function showing how "tenancy.TContextExecutor"
	// can be used in a codebase which implements tenacy.
	exampleFunc := func(exec tenancy.TContextExecutor) {}

	exampleFunc(tx)
}
