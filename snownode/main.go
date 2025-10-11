package snownode

import (
	"math/rand/v2"
	"time"

	"github.com/Dextication/snowflake"
)

var (
	SNOWFLAKE_DATE = time.Date(2024, time.May, 1, 0, 0, 0, 0, time.UTC)
	SNOWFLAKE_EPOCH = SNOWFLAKE_DATE.UnixMilli()
	NODE_ID = rand.Uint32N(1 << 10)
	node *snowflake.Node
)

func init() {
	n, err := snowflake.NewNode(NODE_ID, SNOWFLAKE_DATE, 43, 10, 10)
	if err != nil {
		panic("Can't create snowflake node: " + err.Error())
	}

	node = n
}

func NewID() string {
	return node.Generate().String()
}
