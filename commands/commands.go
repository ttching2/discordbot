package commands

const MangaNotificationString = "manga-notification"
const RoleReactString = "react"
const StrawPollDeadlineString = "strawpoll-deadline"
const TwitterFollowListString = "twitter-follow-list"
const TwitterFollowString = "twitter-follow"
const TwitterUnfollowString = "twitter-unfollow"
const EmojifyString = "emote"

/*
$tourney {link (optional?)} - done
$add_organizer {discord_name} - done
$next_losers_match - done
$ammend_participant {tourney_name} {discord_name} - not rn
$win {optional - participant}  {optional - score format "1-1"} - also sends results to person if specified (done)
$finish_tourney - done
$organizer-list - list organizers
*/
const TournamentCommandString = "tournament"
const TournamentAddOrganizerString = "add-organizer"
const TournamentNextLosersMatchString = "next-losers-match"
const TournamentMatchWinString = "match-win"
const TournamentFinishString = "end-tournament"