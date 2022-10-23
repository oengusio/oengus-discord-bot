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

func assignRoleToRunnersESA(s *discordgo.Session, i *discordgo.InteractionCreate) {
	//assignRolesToRunners(s, "ESA-Win23", EsaDiscord, esaRunnerRole)
	assignRolesToRunners(s, "poggers", EsaDiscord, esaRunnerRole)

	go func() {
		ids, _ := api.GetAcceptedRunnerDiscordIds("poggers")

		for _, id := range ids {
			fakeUser := discordgo.User{
				ID: id,
			}

			s.ChannelMessageSend(i.ChannelID, fakeUser.Mention())
		}
	}()
}
func assignRoleToRunnersBSG(s *discordgo.Session) {
	assignRolesToRunners(s, "", BsgDiscord, bsgRunnerRole)
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
	//      - what channel to log in when audit log is not set? (use command??)

	go startRoleAssignment(s, marathonId, guildId, roleId)
}

func startRoleAssignment(s *discordgo.Session, marathonId, guildId, roleId string) {
	userIDs, err := api.GetAcceptedRunnerDiscordIds("poggers")

	if err != nil {
		fmt.Println("Failed to load users for marathon", marathonId, err)
		return
	}

	for _, userId := range userIDs {
		applyRoleToUser(s, userId, guildId, roleId)
	}

	fmt.Println("Done!")
}

func findMember(s *discordgo.Session, guildId, userId string) *discordgo.Member {
	member, err := s.GuildMember(guildId, userId)

	if err == nil {
		return member
	}

	return nil
}

func applyRoleToUser(s *discordgo.Session, userId, guildId, roleId string) {
	member := findMember(s, guildId, userId)

	if member == nil {
		fmt.Println("Null member for", userId)
		return
	}

	if memberHasRole(member, roleId) {
		fmt.Println(member.User.String(), "already has the role, skipping")
		return
	}

	err := s.GuildMemberRoleAdd(guildId, userId, roleId)

	if err != nil {
		fmt.Println("applying role to", member.User.String(), "failed", err)
	}
}

func removeRoleFromRunners(s *discordgo.Session, marathonId, guildId, roleId string) {
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
