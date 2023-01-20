package discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

func lookupGuild(s *discordgo.Session, guildId string) (*discordgo.Guild, error) {
	guild, err := s.State.Guild(guildId)
	if err != nil {
		guild, err = updateStateGuilds(s, guildId)
		if err != nil {
			return nil, fmt.Errorf("unable to query guild: %w", err)
		}
	}

	return guild, nil
}

// https://github.com/ewohltman/ephemeral-roles/blob/9bc80e6f1111d6959c1053ca2592cee141b94b85/internal/pkg/operations/operations.go#L205

func updateStateGuilds(session *discordgo.Session, guildID string) (*discordgo.Guild, error) {
	guild, err := session.Guild(guildID)
	if err != nil {
		return nil, fmt.Errorf("error sending guild query request: %w", err)
	}

	roles, err := session.GuildRoles(guildID)
	if err != nil {
		return nil, fmt.Errorf("unable to query guild channels: %w", err)
	}

	channels, err := session.GuildChannels(guildID)
	if err != nil {
		return nil, fmt.Errorf("unable to query guild channels: %w", err)
	}

	members, err := recursiveGuildMembers(session, guildID, "", guildMembersPageLimit)
	if err != nil {
		return nil, fmt.Errorf("unable to query guild members: %w", err)
	}

	guild.Roles = roles
	guild.Channels = channels
	guild.Members = members
	guild.MemberCount = len(members)

	err = session.State.GuildAdd(guild)
	if err != nil {
		return nil, fmt.Errorf("unable to add guild to state cache: %w", err)
	}

	return guild, nil
}

func recursiveGuildMembers(
	session *discordgo.Session,
	guildID, after string,
	limit int,
) ([]*discordgo.Member, error) {
	guildMembers, err := session.GuildMembers(guildID, after, limit)
	if err != nil {
		return nil, fmt.Errorf("error sending recursive guild members request: %w", err)
	}

	if len(guildMembers) < guildMembersPageLimit {
		return guildMembers, nil
	}

	nextGuildMembers, err := recursiveGuildMembers(
		session,
		guildID,
		guildMembers[len(guildMembers)-1].User.ID,
		guildMembersPageLimit,
	)
	if err != nil {
		return nil, err
	}

	return append(guildMembers, nextGuildMembers...), nil
}
