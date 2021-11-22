package model

import (
	"database/sql"
)

type ID = uint32

type IdentifiableEntity struct {
	ID ID
}

type Snowflake = uint64

type IdentifiableDiscordEntity struct {
	IdentifiableEntity
	DiscordID Snowflake
}

type NullableSnowflake = sql.NullInt64

type NullIdentifiableDiscordEntity struct {
	IdentifiableEntity
	DiscordID NullableSnowflake
}

type Ref = uint32
