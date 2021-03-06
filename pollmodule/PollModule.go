package pollmodule

import (
	"fmt"
	"strconv"
	"strings"

	bot "../sweetiebot"
	"github.com/blackhole12/discordgo"
)

// PollModule manages the polling system
type PollModule struct {
}

// New PollModule
func New() *PollModule {
	return &PollModule{}
}

// Name of the module
func (w *PollModule) Name() string {
	return "Polls"
}

// Commands in the module
func (w *PollModule) Commands() []bot.Command {
	return []bot.Command{
		&pollCommand{},
		&createPollCommand{},
		&deletePollCommand{},
		&voteCommand{},
		&resultsCommand{},
		&addOptionCommand{},
	}
}

// Description of the module
func (w *PollModule) Description() string { return "Manages the polling system." }

type pollCommand struct {
}

func (c *pollCommand) Info() *bot.CommandInfo {
	return &bot.CommandInfo{
		Name:  "Poll",
		Usage: "Displays poll description and options.",
	}
}
func (c *pollCommand) Process(args []string, msg *discordgo.Message, indices []int, info *bot.GuildInfo) (string, bool, *discordgo.MessageEmbed) {
	if !info.Bot.DB.CheckStatus() {
		return "```\nA temporary database outage is preventing this command from being executed.```", false, nil
	}
	gID := bot.SBatoi(info.ID)
	if len(args) < 1 {
		polls := info.Bot.DB.GetPolls(gID)
		str := make([]string, 0, len(polls)+1)
		str = append(str, "All active polls:")

		for _, v := range polls {
			str = append(str, v.Name)
		}
		return "```\n" + strings.Join(str, "\n") + "```", len(str) > bot.MaxPublicLines, nil
	}
	arg := strings.ToLower(msg.Content[indices[0]:])
	id, desc := info.Bot.DB.GetPoll(arg, gID)
	if id == 0 {
		return "```\nThat poll doesn't exist!```", false, nil
	}
	options := info.Bot.DB.GetOptions(id)

	str := make([]string, 0, len(options)+2)
	str = append(str, desc)

	for _, v := range options {
		str = append(str, fmt.Sprintf("%v. %s", v.Index, v.Option))
	}

	return strings.Join(str, "\n"), len(str) > bot.MaxPublicLines, nil
}
func (c *pollCommand) Usage(info *bot.GuildInfo) *bot.CommandUsage {
	return &bot.CommandUsage{
		Desc: "Displays currently active polls or possible options for a given poll.",
		Params: []bot.CommandUsageParam{
			{Name: "poll", Desc: "Name of a specific poll to display.", Optional: true},
		},
	}
}

type createPollCommand struct {
}

func (c *createPollCommand) Info() *bot.CommandInfo {
	return &bot.CommandInfo{
		Name:      "CreatePoll",
		Usage:     "Creates a poll.",
		Sensitive: true,
	}
}
func (c *createPollCommand) Process(args []string, msg *discordgo.Message, indices []int, info *bot.GuildInfo) (string, bool, *discordgo.MessageEmbed) {
	if !info.Bot.DB.CheckStatus() {
		return "```\nA temporary database outage is preventing this command from being executed.```", false, nil
	}
	if len(args) < 3 {
		return "```\nYou must provide a name, a description, and one or more options to create the poll. Example: " + info.Config.Basic.CommandPrefix + "createpoll pollname \"Description With Space\" \"Option 1\" \"Option 2\"```", false, nil
	}
	gID := bot.SBatoi(info.ID)
	name := strings.ToLower(args[0])
	err := info.Bot.DB.AddPoll(name, args[1], gID)
	if err == bot.ErrDuplicateEntry {
		return "```\nError, poll name already used.```", false, nil
	} else if err != nil {
		return bot.ReturnError(err)
	}
	poll, _ := info.Bot.DB.GetPoll(name, gID)
	if poll == 0 {
		return "```\nError: Orphaned poll!```", false, nil
	}

	for k, v := range args[2:] {
		err = info.Bot.DB.AddOption(poll, uint64(k+1), v)
		if err == bot.ErrDuplicateEntry {
			return fmt.Sprintf("```\nError adding duplicate option %v:%s. Each option must be unique!```", k+1, v), false, nil
		} else if err != nil {
			return bot.ReturnError(err)
		}
	}

	return fmt.Sprintf("```\nSuccessfully created %s poll.```", name), false, nil
}
func (c *createPollCommand) Usage(info *bot.GuildInfo) *bot.CommandUsage {
	return &bot.CommandUsage{
		Desc: "Creates a new poll with the given name, description, and options. All arguments MUST use quotes if they have spaces. \n\nExample usage: `" + info.Config.Basic.CommandPrefix + "createpoll pollname \"Description With Space\" \"Option 1\" NoSpaceOption`",
		Params: []bot.CommandUsageParam{
			{Name: "name", Desc: "Name of the new poll. It's suggested to not use spaces because this makes things difficult for other commands. ", Optional: false},
			{Name: "description", Desc: "Poll description that appears when displaying it.", Optional: false},
			{Name: "options", Desc: "Name of the new poll. It's suggested to not use spaces because this makes things difficult for other commands. ", Optional: true, Variadic: true},
		},
	}
}

type deletePollCommand struct {
}

func (c *deletePollCommand) Info() *bot.CommandInfo {
	return &bot.CommandInfo{
		Name:      "DeletePoll",
		Usage:     "Deletes a poll.",
		Sensitive: true,
	}
}
func (c *deletePollCommand) Process(args []string, msg *discordgo.Message, indices []int, info *bot.GuildInfo) (string, bool, *discordgo.MessageEmbed) {
	if !info.Bot.DB.CheckStatus() {
		return "```\nA temporary database outage is preventing this command from being executed.```", false, nil
	}
	if len(args) < 1 {
		return "```\nYou have to give me a poll name to delete!```", false, nil
	}
	arg := msg.Content[indices[0]:]
	gID := bot.SBatoi(info.ID)
	id, _ := info.Bot.DB.GetPoll(arg, gID)
	if id == 0 {
		return "```\nThat poll doesn't exist!```", false, nil
	}
	err := info.Bot.DB.RemovePoll(arg, gID)
	if err != nil {
		return bot.ReturnError(err)
	}
	return fmt.Sprintf("```\nSuccessfully removed %s.```", arg), false, nil
}
func (c *deletePollCommand) Usage(info *bot.GuildInfo) *bot.CommandUsage {
	return &bot.CommandUsage{
		Desc: "Removes the poll with the given poll name.",
		Params: []bot.CommandUsageParam{
			{Name: "poll", Desc: "Name of the poll to delete.", Optional: false},
		},
	}
}

type voteCommand struct {
}

func (c *voteCommand) Info() *bot.CommandInfo {
	return &bot.CommandInfo{
		Name:  "Vote",
		Usage: "Votes in a poll.",
	}
}
func (c *voteCommand) Process(args []string, msg *discordgo.Message, indices []int, info *bot.GuildInfo) (string, bool, *discordgo.MessageEmbed) {
	if !info.Bot.DB.CheckStatus() {
		return "```\nA temporary database outage is preventing this command from being executed.```", false, nil
	}
	gID := bot.SBatoi(info.ID)
	if len(args) < 2 {
		polls := info.Bot.DB.GetPolls(gID)
		lastpoll := ""
		if len(polls) > 0 {
			lastpoll = fmt.Sprintf(" The most recent poll is \"%s\".", polls[0].Name)
		}
		return fmt.Sprintf("```\nYou have to provide both a poll name and the option you want to vote for!%s Use "+info.Config.Basic.CommandPrefix+"poll without any arguments to list all active polls.```", lastpoll), false, nil
	}
	name := strings.ToLower(args[0])
	id, _ := info.Bot.DB.GetPoll(name, gID)
	if id == 0 {
		return "```\nThat poll doesn't exist! Use " + info.Config.Basic.CommandPrefix + "poll with no arguments to list all active polls.```", false, nil
	}

	option, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		opt := info.Bot.DB.GetOption(id, msg.Content[indices[1]:])
		if opt == nil {
			return fmt.Sprintf("```\nThat's not one of the poll options! You have to either type in the exact name of the option you want, or provide the numeric index. Use \""+info.Config.Basic.CommandPrefix+"poll %s\" to list the available options.```", name), false, nil
		}
		option = *opt
	} else if !info.Bot.DB.CheckOption(id, option) {
		return fmt.Sprintf("```\nThat's not a valid option index! Use \""+info.Config.Basic.CommandPrefix+"poll %s\" to get all available options for this poll.```", name), false, nil
	}

	err = info.Bot.DB.AddVote(bot.SBatoi(msg.Author.ID), id, option)
	if err != nil {
		return bot.ReturnError(err)
	}

	return "```\nVoted! Use " + info.Config.Basic.CommandPrefix + "results to check the results.```", false, nil
}
func (c *voteCommand) Usage(info *bot.GuildInfo) *bot.CommandUsage {
	return &bot.CommandUsage{
		Desc: "Adds your vote to a given poll. If you have already voted in the poll, it changes your vote instead.",
		Params: []bot.CommandUsageParam{
			{Name: "poll", Desc: "Name of the poll you want to vote in.", Optional: false},
			{Name: "option", Desc: "The numeric index of the option you want to vote for, or the precise text of the option instead.", Optional: false},
		},
	}
}

type resultsCommand struct {
}

func (c *resultsCommand) Info() *bot.CommandInfo {
	return &bot.CommandInfo{
		Name:  "Results",
		Usage: "Gets poll results.",
	}
}

func (c *resultsCommand) Process(args []string, msg *discordgo.Message, indices []int, info *bot.GuildInfo) (string, bool, *discordgo.MessageEmbed) {
	if !info.Bot.DB.CheckStatus() {
		return "```\nA temporary database outage is preventing this command from being executed.```", false, nil
	}
	gID := bot.SBatoi(info.ID)
	if len(args) < 1 {
		return "```\nYou have to give me a valid poll name! Use \"" + info.Config.Basic.CommandPrefix + "poll\" to list active polls.```", false, nil
	}
	arg := strings.ToLower(msg.Content[indices[0]:])
	id, desc := info.Bot.DB.GetPoll(arg, gID)
	if id == 0 {
		return "```\nThat poll doesn't exist! Use \"" + info.Config.Basic.CommandPrefix + "poll\" to list active polls.```", false, nil
	}
	results := info.Bot.DB.GetResults(id)
	options := info.Bot.DB.GetOptions(id)
	max := uint64(0)
	for _, v := range results {
		if v.Count > max {
			max = v.Count
		}
	}

	str := make([]string, 0, len(results)+2)
	str = append(str, desc)
	k := 0
	var count uint64
	for _, v := range options {
		count = 0
		if k < len(results) && v.Index == results[k].Index {
			count = results[k].Count
			k++
		}
		normalized := count
		if max > 10 {
			normalized = uint64(float32(count) * (10.0 / float32(max)))
		}
		if count > 0 && normalized < 1 {
			normalized = 1
		}

		graph := ""
		for i := 0; i < 10; i++ {
			if uint64(i) < normalized {
				graph += "\u2588" // this isn't very efficient but the maximum is 10 so it doesn't matter
			} else {
				graph += "\u2591"
			}
		}
		buf := ""
		if v.Index < 10 && len(options) > 9 {
			buf = "_"
		}
		str = append(str, fmt.Sprintf("`%s%v. `%s %s (%v votes)", buf, v.Index, graph, v.Option, count))
	}

	return strings.Join(str, "\n"), len(str) > 11, nil
}
func (c *resultsCommand) Usage(info *bot.GuildInfo) *bot.CommandUsage {
	return &bot.CommandUsage{
		Desc: "Displays the results of the given poll, if it exists.",
		Params: []bot.CommandUsageParam{
			{Name: "poll", Desc: "Name of the poll to view.", Optional: false},
		},
	}
}

type addOptionCommand struct {
}

func (c *addOptionCommand) Info() *bot.CommandInfo {
	return &bot.CommandInfo{
		Name:      "AddOption",
		Usage:     "Appends an option to a poll.",
		Sensitive: true,
	}
}

func (c *addOptionCommand) Process(args []string, msg *discordgo.Message, indices []int, info *bot.GuildInfo) (string, bool, *discordgo.MessageEmbed) {
	if !info.Bot.DB.CheckStatus() {
		return "```\nA temporary database outage is preventing this command from being executed.```", false, nil
	}
	if len(args) < 1 {
		return "```\nYou have to give me a poll name to add an option to!```", false, nil
	}
	if len(args) < 2 {
		return "```\nYou have to give me an option to add!```", false, nil
	}
	gID := bot.SBatoi(info.ID)
	id, _ := info.Bot.DB.GetPoll(args[0], gID)
	if id == 0 {
		return "```\nThat poll doesn't exist!```", false, nil
	}
	arg := msg.Content[indices[1]:]
	err := info.Bot.DB.AppendOption(id, arg)
	if err == bot.ErrDuplicateEntry {
		return "```\nAnother option is already called this! Options must be unique.```", false, nil
	} else if err != nil {
		return bot.ReturnError(err)
	}
	return fmt.Sprintf("```\nSuccessfully added %s to %s.```", arg, args[0]), false, nil
}
func (c *addOptionCommand) Usage(info *bot.GuildInfo) *bot.CommandUsage {
	return &bot.CommandUsage{
		Desc: "Appends an option to a poll.",
		Params: []bot.CommandUsageParam{
			{Name: "poll", Desc: "Name of the poll to modify.", Optional: false},
			{Name: "option", Desc: "The option to append to the end of the poll.", Optional: false},
		},
	}
}
