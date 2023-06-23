package discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/exp/slices"
	"oenugs-bot/api"
)

var guildMembersPageLimit = 1000

// TODO:
//  1. test that role exists
//  2. Test that bot is in server
//  3. Send updates when members are not in the server/assignment failed

func AssignRoleToRunners(s *discordgo.Session, i *discordgo.InteractionCreate, marathonId, guildId, roleId string) {
	assignRolesToRunners(s, i, marathonId, i.ChannelID, guildId, roleId)
}

func RemoveRolesFromRunners(s *discordgo.Session, i *discordgo.InteractionCreate, marathonId, guildId, roleId string) {
	go func() {
		selectionDone, err := api.GetMarathonSelectionDone(marathonId)

		if err != nil {
			s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
				Content: "Error: " + err.Error(),
			})
			return
		}

		if !selectionDone {
			s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
				Content: "Marathon has not completed selection yet???",
			})
			return
		}

		removeRoleFromRunners(s, marathonId, i.ChannelID, guildId, roleId)
	}()
}

func memberHasRole(member *discordgo.Member, roleId string) bool {
	return slices.Contains(member.Roles, roleId)
}

func assignRolesToRunners(s *discordgo.Session, i *discordgo.InteractionCreate, marathonId, channelId, guildId, roleId string) {
	// TODO
	//  1. Fetch marathon settings (already done before this func)
	//  2. Fetch submissions with status: VALIDATED, BACKUP, BONUS
	//  3. Assign role
	//  4. log feedback about runners that could not be assigned a role with reason (no perms or no discord id)
	//      - Will be activated via command

	go func() {
		selectionDone, err := api.GetMarathonSelectionDone(marathonId)

		if err != nil {
			s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
				Content: "Error: " + err.Error(),
			})
			return
		}

		if !selectionDone {
			s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
				Content: "Marathon has not completed selection yet.",
			})
			return
		}

		startRoleAssignment(s, marathonId, channelId, guildId, roleId)
	}()
}

func startRoleAssignment(s *discordgo.Session, marathonId, channelId, guildId, roleId string) {
	userIDs, err := api.GetAcceptedRunnerDiscordIds(marathonId)

	if err != nil {
		fmt.Println("Failed to load users for marathon", marathonId, err)
		return
	}

	for _, userId := range userIDs {
		applyRoleToUser(s, userId, guildId, roleId)
	}

	fmt.Println("Done!")
	s.ChannelMessageSend(channelId, "Role assignment has been completed!")
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

func removeRoleFromRunners(s *discordgo.Session, marathonId, channelId, guildId, roleId string) {
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

		s.ChannelMessageSend(channelId, "Role removal has been completed!")
	}()
}
