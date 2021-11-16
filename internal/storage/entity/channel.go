package entity

type Channel struct {
	IdentifiableDiscordEntity
	GuildID Snowflake
}

func NewChannel(ID ID, discordID Snowflake, guildID Snowflake) *Channel {
	return &Channel{IdentifiableDiscordEntity{IdentifiableEntity{ID}, discordID}, guildID}
}