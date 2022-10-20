package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"oenugs-bot/api"
)

var EsaDiscord = "85369684286767104"
var esaRunnerRole = "1015153036064202772"

// BSG also wants to test
var BsgDiscord = "153911811232497664"
var bsgRunnerRole = "630687834688323594"

var botTestingChannel = "798952970892214272"

var guildMembersPageLimit = 1000

// TODO: binary search the shit out of this

func assignRoleToRunnersESA(s *discordgo.Session) {
	//assignRolesToRunners(s, "ESA-Win23", EsaDiscord, esaRunnerRole)
	assignRolesToRunners(s, "caching", EsaDiscord, esaRunnerRole)
}
func assignRoleToRunnersBSG(s *discordgo.Session) {
	assignRolesToRunners(s, "", BsgDiscord, bsgRunnerRole)
}

func removeRoleFromRunners(s *discordgo.Session, marathonId string, guildId string, roleId string) {
	guild, err := lookupGuild(s, guildId)

	if err != nil {
		fmt.Println("Failed to look up guild for marathon ", marathonId, guildId, err)
		return
	}

	go func() {
		if guild.MemberCount > len(guild.Members) {
			members, err := recursiveGuildMembers(s, guild.ID, "", guildMembersPageLimit)
			if err != nil {
				fmt.Printf("unable to query guild members: %s\n", err)
				return
			}

			guild.Members = members
			guild.MemberCount = len(members)
		}

		for _, member := range guild.Members {
			if memberHasRole(member, roleId) {
				err := s.GuildMemberRoleRemove(guild.ID, member.User.ID, roleId)

				if err != nil {
					fmt.Println("Failed to remove role from", member.User.String(), err)
					return
				}
			}
		}
	}()
}

func memberHasRole(member *discordgo.Member, roleId string) bool {
	for _, mRole := range member.Roles {
		if mRole == roleId {
			return true
		}
	}

	return false
}

func assignRolesToRunners(s *discordgo.Session, marathonId string, guildId string, roleId string) {
	// TODO
	//  1. Fetch marathon settings (already done before this func)
	//  2. Fetch submissions with status: VALIDATED, BACKUP, BONUS
	//  3. Assign role
	//  4. log feedback (audit channel??) about runners that could not be assigned a role with reason (no perms or no discord id)
	//      - what channel to log in when audit log is not set?

	guild, err := lookupGuild(s, guildId)

	if err != nil {
		fmt.Println("Failed to look up guild for marathon ", marathonId, guildId, err)
		return
	}

	fmt.Println(guild.Members, len(guild.Members))

	go startRoleAssignment(marathonId, s, guild, roleId)
}

// TODO: look up user id with tag from member cache

func startRoleAssignment(marathonId string, s *discordgo.Session, guild *discordgo.Guild, roleId string) {
	submissions, err := api.FetchAcceptedRunners(marathonId)

	if err != nil {
		s.ChannelMessageSend(botTestingChannel, err.Error())
		return
	}

	if guild.MemberCount > len(guild.Members) {
		members, err := recursiveGuildMembers(s, guild.ID, "", guildMembersPageLimit)
		if err != nil {
			fmt.Printf("unable to query guild members: %s\n", err)
			return
		}

		guild.Members = members
		guild.MemberCount = len(members)
	}

	for _, submission := range submissions {
		discordTag := findDiscordUsername(submission)

		if discordTag == "" {
			fmt.Printf("No discord for %s\n", submission.User.Username)
			continue
		}

		member := findUser(discordTag, guild.Members)

		if member == nil {
			fmt.Printf("%s is not in guild (correct discord linked?)\n", discordTag)
			continue
		}

		applyRoleToUser(s, member, guild.ID, roleId)
	}

	fmt.Println("Done!")
}

func findDiscordUsername(sub api.Submission) string {
	for _, cnx := range sub.User.Connections {
		if cnx.Platform == "DISCORD" {
			return cnx.Username
		}
	}

	return ""
}

func findUser(tag string, members []*discordgo.Member) *discordgo.Member {
	for _, member := range members {
		if member.User.String() == tag {
			return member
		}
	}

	return nil
}

func applyRoleToUser(s *discordgo.Session, member *discordgo.Member, guildId, roleId string) {
	if memberHasRole(member, roleId) {
		fmt.Println(member.User.String(), "already has the role, skipping")
		return
	}

	err := s.GuildMemberRoleAdd(guildId, member.User.ID, roleId)

	if err != nil {
		fmt.Println("applying role to", member.User.String(), "failed", err)
	}
}
