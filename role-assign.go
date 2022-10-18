package main

var esaDiscord = "85369684286767104"
var esaRunnerRole = "1015153036064202772"

// BSG also wants to test
var bsgDiscord = "153911811232497664"
var bsgRunnerRole = "630687834688323594"

func assignRoleToRunnersESA() {
	assignRolesToRunners("ESA-Win23", esaDiscord, esaRunnerRole)
}
func assignRoleToRunnersBSG() {
	assignRolesToRunners("", bsgDiscord, bsgRunnerRole)
}

func assignRolesToRunners(marathonId string, guildId string, roleId string) {
	// TODO
	//  1. Fetch marathon settings (already done before this func)
	//  2. Fetch submissions with status: VALIDATED, BACKUP, BONUS
	//  3. Assign role
	//  4. log feedback (audit channel??) about runners that could not be assigned a role with reason (no perms or no discord id)
	//      - what channel to log in when audit log is not set?
}
