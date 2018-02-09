package miscmodule

import (
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	bot "../sweetiebot"
	"github.com/blackhole12/discordgo"
)

var diceregex = regexp.MustCompile("[0-9]*d[0-9]+")

type showrollCommand struct {
}

func (c *showrollCommand) Info() *bot.CommandInfo {
	return &bot.CommandInfo{
		Name:              "Showroll",
		Usage:             "Evaluates a dice expression.",
		ServerIndependent: true,
	}
}
func (c *showrollCommand) eval(args []string, index *int, info *bot.GuildInfo) string {
	var s string
	for *index < len(args) {
		s += value(args, index) + "\n"
	}
	return s
}
func (c *showrollCommand) value(args []string, index *int, info *bot.GuildInfo) string {
	*index++
	if diceregex.MatchString(args[*index-1]) {
		dice := strings.SplitN(args[*index-1], "d", 2)
		var multiplier, num, threshold, fail int64 = 1, 1, 0, 0
		var s string
		s = "Rolling " + args[*index-1] + ": "
		if len(dice) > 1 {
			if len(dice[0]) > 0 {
				multiplier, _ = strconv.ParseInt(dice[0], 10, 64)
			}
			if strings.Contains(dice[1], "t") {
				tdice := strings.SplitN(dice[1], "t", 2)
				dice[1] = tdice[0]
				if strings.Contains(tdice[1], "f") {
					fdice := strings.SplitN(tdice[1], "f", 2)
					threshold, _ = strconv.ParseInt(fdice[0], 10, 64)
					fail, _ = strconv.ParseInt(fdice[1], 10, 64)
				} else {
					threshold, _ = strconv.ParseInt(tdice[1], 10, 64)
				}
				if threshold == 0 {
					return s + "Can't roll that! Check dice expression."
				}
			}
			if strings.Contains(dice[1], "f") {
				fdice := strings.SplitN(dice[1], "f", 2)
				dice[1] = fdice[0]
				fail, _ = strconv.ParseInt(fdice[1], 10, 64)
				if fail == 0 {
					return s + "Can't roll that! Check dice expression."
				}
			}
			num, _ = strconv.ParseInt(dice[1], 10, 64)
		} else {
			num, _ = strconv.ParseInt(dice[0], 10, 64)
		}
		if multiplier < 1 || num < 1 {
			return s + "Can't roll that! Check dice expression."
		}
		if multiplier > 9999 {
			return s + "I don't have that many dice..."
		}
		var n int64
		var t int = 0
		var f int = 0
		for ; multiplier > 0; multiplier-- {
			n = rand.Int63n(num) + 1
			s += strconv.FormatInt(n, 10)
			if multiplier > 1 {
				s += " + "
			}
			if threshold > 0 {
				if n >= threshold {
					t++
				}
			}
			if fail > 0 {
				if n <= fail {
					f++
				}
			}
		}
		if t > 0 {
			s += "\n" + strconv.Itoa(t) + " successes!"
		}
		if f > 0 {
			s += "\n" + strconv.Itoa(f) + " failures!"
		}
		return s
	}
	return "Could not parse dice expression. Try " + info.Config.Basic.CommandPrefix + "calculate for advanced expressions."
}
func (c *showrollCommand) Process(args []string, msg *discordgo.Message, indices []int, info *bot.GuildInfo) (retval string, b bool, embed *discordgo.MessageEmbed) {
	if len(args) < 1 {
		return "```\nNothing to roll!```", false, nil
	}
	defer func() {
		if s := recover(); s != nil {
			retval = "```ERROR: " + s.(string) + "```"
		}
	}()
	index := 0
	s := c.eval(args, &index, info)
	return "```\n" + s + "```", false, nil
}
func (c *showrollCommand) Usage(info *bot.GuildInfo) *bot.CommandUsage {
	return &bot.CommandUsage{
		Desc: "Evaluates a dice roll expression (**N**d**X**[**t**X**][**f**X**]), returning the individual die results. For example, `" + info.Config.Basic.CommandPrefix + "showroll d10` will return `5`, whereas `" + info.Config.Basic.CommandPrefix + "showroll 4d6` will return: `6 + 4 + 2 + 3`. Specify success and failure thresholds with tx and fx respectively. e.g. `17d6t5f2` will report the number of dice 5 or above (successes) and 2 or below (failures)",
		Params: []bot.CommandUsageParam{
			{Name: "expression", Desc: "The dice expression to parse.", Optional: false},
		},
	}
}