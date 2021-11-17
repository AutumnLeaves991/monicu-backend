package entity

import (
	"fmt"
	"strconv"
)

func mustParseSnowflake(s string) Snowflake {
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		panic(fmt.Errorf("could not parse Snowflake ID string: %w", err))
	}
	return val
}
