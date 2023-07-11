package slashHandlers

import (
	"github.com/bwmarrin/discordgo"
	"golang.org/x/exp/slices"
	"oenugs-bot/api"
	"oenugs-bot/discord"
	"oenugs-bot/utils"
)

func HandleRoleManagement(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Member == nil {
		return
	}

	data := i.ApplicationCommandData()

	// This should not be possible, but just to be safe.
	if len(data.Options) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "Please choose assign or remove",
			},
		})
		return
	}

	subCmd := data.Options[0]
	subName := subCmd.Name

	if subName != "assign" && subName != "remove" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "Either discord done fucked up or you need to report a bug",
			},
		})
		return
	}

	// Don't want to rely on the order of the options
	options := utils.OptionsToMap(subCmd.Options)
	marathonId := options["marathon"].StringValue()
	moderators, err := api.GetModeratorsForMarathon(marathonId)

	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "Failed to look up moderators for marathon, try again later: " + err.Error(),
			},
		})
		return
	}

	if !slices.Contains(moderators, i.Member.User.ID) {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "You must be a marathon owner or moderator to run this command",
			},
		})
		return
	}

	//guild, _ := s.State.Gu\ild(i.GuildID)
	//selfPerms := guild.Permissions

	botHasPerms, _ := utils.MemberHasPermission(s, i.GuildID, i.AppID, discordgo.PermissionManageRoles)

	if !botHasPerms {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "I need the manage roles permission in order to give roles to people.",
			},
		})
		return
	}

	// Never returns null
	role := options["role"].RoleValue(s, i.GuildID)
	roleId := role.ID
	selfMem, _ := s.GuildMember(i.GuildID, i.AppID) // This should not error LOL
	maxRoleIdSelfMem := selfMem.Roles[0]
	maxRoleSelfMem, _ := s.State.Role(i.GuildID, maxRoleIdSelfMem)

	if role.Position > maxRoleSelfMem.Position {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "I cannot interact with " + role.Mention() + ", please make sure that it is below my own bot role",
			},
		})
		return
	}

	switch subCmd.Name {
	case "assign":
		discord.AssignRoleToRunners(s, i, marathonId, i.GuildID, roleId)

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Runner role assignment started",
			},
		})
		break
	case "remove":
		discord.RemoveRolesFromRunners(s, i, marathonId, i.GuildID, roleId)

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Runner role removal started",
			},
		})
		break
	default:
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "I'm impressed, please report this bug",
			},
		})
		break
	}
}
