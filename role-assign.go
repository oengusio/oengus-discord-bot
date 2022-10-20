package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

var EsaDiscord = "85369684286767104"
var esaRunnerRole = "1015153036064202772"

// BSG also wants to test
var BsgDiscord = "153911811232497664"
var bsgRunnerRole = "630687834688323594"

func assignRoleToRunnersESA(s *discordgo.Session) {
	assignRolesToRunners(s, "ESA-Win23", EsaDiscord, esaRunnerRole)
}
func assignRoleToRunnersBSG(s *discordgo.Session) {
	assignRolesToRunners(s, "", BsgDiscord, bsgRunnerRole)
}

func assignRolesToRunners(s *discordgo.Session, marathonId string, guildId string, roleId string) {
	// TODO
	//  1. Fetch marathon settings (already done before this func)
	//  2. Fetch submissions with status: VALIDATED, BACKUP, BONUS
	//  3. Assign role
	//  4. log feedback (audit channel??) about runners that could not be assigned a role with reason (no perms or no discord id)
	//      - what channel to log in when audit log is not set?

	guild, err := s.Guild(guildId)

	if err != nil {
		fmt.Println("Failed to look up guild for marathon ", marathonId, guildId, err)
		return
	}

	var role *discordgo.Role = nil

	for _, listRole := range guild.Roles {
		if listRole.ID == roleId {
			role = listRole
			break
		}
	}

	if role == nil {
		fmt.Println("Role is nil")
		return
	}

	fmt.Println("Our permissions", marathonId, guild.Roles)
}

// TODO: look up user id with tag from member cache

func startRoleAssignment(s *discordgo.Session, guild *discordgo.Guild, role *discordgo.Role) {
	//api.FetchAcceptedRunners()
}

func applyRoleToUser(s *discordgo.Session, userId string, guild *discordgo.Guild, role *discordgo.Role) {
	//
}
