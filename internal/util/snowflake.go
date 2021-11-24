package util

import (
	"fmt"
	"strconv"
)

func MustParseSnowflake(s string) uint64 {
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		panic(fmt.Errorf("could not parse Snowflake ID string: %w", err))
	}
	return val
}

func MustParseSnowflakeInt64(s string) int64 {
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(fmt.Errorf("could not parse Snowflake ID string: %w", err))
	}
	return val
}

func FormatSnowflake(s uint64) string {
	return strconv.FormatUint(s, 10)
}
