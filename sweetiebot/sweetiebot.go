package sweetiebot

import (
  "fmt"
  "time"
  "strconv"
  "sync/atomic"
  "io/ioutil"
  "github.com/bwmarrin/discordgo"
)

type ModuleHooks struct {
    OnEvent                   []ModuleOnEvent
    OnTypingStart             []ModuleOnTypingStart
    OnMessageCreate           []ModuleOnMessageCreate
    OnMessageUpdate           []ModuleOnMessageUpdate
    OnMessageDelete           []ModuleOnMessageDelete
    OnMessageAck              []ModuleOnMessageAck
    OnUserUpdate              []ModuleOnUserUpdate
    OnPresenceUpdate          []ModuleOnPresenceUpdate
    OnVoiceStateUpdate        []ModuleOnVoiceStateUpdate
    OnGuildUpdate             []ModuleOnGuildUpdate
    OnGuildMemberAdd          []ModuleOnGuildMemberAdd
    OnGuildMemberRemove       []ModuleOnGuildMemberRemove
    OnGuildMemberUpdate       []ModuleOnGuildMemberUpdate
    OnGuildBanAdd             []ModuleOnGuildBanAdd
    OnGuildBanRemove          []ModuleOnGuildBanRemove
    OnEvent_channels          []map[uint64]bool
    OnTypingStart_channels    []map[uint64]bool
    OnMessageCreate_channels  []map[uint64]bool
    OnMessageUpdate_channels  []map[uint64]bool
    OnMessageDelete_channels  []map[uint64]bool
    OnMessageAck_channels     []map[uint64]bool
    OnUserUpdate_channels     []map[uint64]bool
    OnPresenceUpdate_channels []map[uint64]bool
    OnVoiceStateUpdate_channels []map[uint64]bool
    OnGuildUpdate_channels    []map[uint64]bool
    OnGuildMemberAdd_channels []map[uint64]bool
    OnGuildMemberRemove_channels []map[uint64]bool
    OnGuildMemberUpdate_channels []map[uint64]bool
    OnGuildBanAdd_channels    []map[uint64]bool
    OnGuildBanRemove_channels []map[uint64]bool
}

type SweetieBot struct {
  db *BotDB
  log *Log
  dg *discordgo.Session
  SelfID string
  GuildID string
  LogChannelID string
  ModChannelID string
  version string
  debug bool
  hooks ModuleHooks
  modules []Module
  commands []Command
}

var sb *SweetieBot

func RateLimit(prevtime *int64, interval int64) bool {
  t := time.Now().UTC().Unix()
  d := *prevtime // perform a read so it doesn't change on us
  if t - d > interval {
    return atomic.CompareAndSwapInt64(prevtime, d, t) // If the swapped failed, it means another thread already sent a message and swapped it out, so don't send a message.
  } 
  return false
}

func SBatoi(s string) uint64 {
  i, err := strconv.ParseUint(s, 10, 64)
  if err != nil { 
    sb.log.Log("Invalid number ", s)
    return 0 
  }
  return i
}

func ProcessModules(channels []map[uint64]bool, channelID string, fn func(i int)) {
  if len(channels)>0 { // only bother doing anything if we actually have hooks to process
    for i, c := range channels {
      if len(channelID)>0 && len(c)>0 { // Only check for channels if we have a channel to check for, and the module actually has specific channels
        _, ok := c[SBatoi(channelID)]
        if !ok { continue; }
      }
      fn(i)
    }
  }
}

func SBEvent(s *discordgo.Session, e *discordgo.Event) { ProcessModules(sb.hooks.OnEvent_channels, "", func(i int) { sb.hooks.OnEvent[i].OnEvent(s, e) }) }
func SBReady(s *discordgo.Session, r *discordgo.Ready) {
  fmt.Println("Ready message receieved")
  sb.SelfID = r.User.ID
  g := r.Guilds[0]
  ProcessGuild(g)
  
  for _, v := range g.Members {
    ProcessMember(v)
  }
  
  modules := ""
  commands := ""
  
  for _, v := range sb.modules {
    modules += "\n  "
    modules += v.Name() 
  }
  for _, v := range sb.commands {
    commands += "\n  "
    commands += v.Name() 
  }
    
  sb.log.Log("[](/sbload) Sweetiebot version ", sb.version, " successfully loaded on ", g.Name, ". \nActive Modules:", modules, "\n\nActive Commands:", commands);
}
func SBTypingStart(s *discordgo.Session, t *discordgo.TypingStart) { ProcessModules(sb.hooks.OnTypingStart_channels, "", func(i int) { sb.hooks.OnTypingStart[i].OnTypingStart(s, t) }) }
func SBMessageCreate(s *discordgo.Session, m *discordgo.Message) {
  if m.Author == nil { // This shouldn't ever happen but we check for it anyway
    return
  }
	fmt.Printf("[%s] %20s %20s %s (%s:%s) > %s\n", m.ID, m.ChannelID, m.Timestamp, m.Author.Username, m.Author.ID, m.Author.Email, m.Content); // DEBUG
  
  if m.ChannelID != sb.LogChannelID { // Log this message provided it wasn't sent to the bot-log channel.
    sb.db.AddMessage(SBatoi(m.ID), SBatoi(m.Author.ID), m.ContentWithMentionsReplaced(), SBatoi(m.ChannelID), m.MentionEveryone) 
  }
  if m.Author.ID == sb.SelfID { // ALWAYS discard any of our own messages before analysis.
    return
  }
  
  if !sb.debug && m.ChannelID == sb.LogChannelID { // Discard any messages sent in the bot-log channel if we aren't in debug mode.
    return
  }
  
  // Check if this is a command. If it is, process it as a command, otherwise process it with our modules.
  if len(m.Content) > 0 && m.Content[0] == '!' {
    // if we've his the saturation limit, post an error (which itself will only post if the error saturation limit hasn't been hit)
  } else {
    ProcessModules(sb.hooks.OnMessageCreate_channels, m.ChannelID, func(i int) { sb.hooks.OnMessageCreate[i].OnMessageCreate(s, m) })  
  }  
}

func SBMessageUpdate(s *discordgo.Session, m *discordgo.Message) {
  if m.Author == nil { // Discord sends an update message with an empty author when certain media links are posted
    return
  }
  if m.ChannelID != sb.LogChannelID { // Always ignore messages from the log channel
    sb.db.AddMessage(SBatoi(m.ID), SBatoi(m.Author.ID), m.ContentWithMentionsReplaced(), SBatoi(m.ChannelID), m.MentionEveryone) 
  }
  ProcessModules(sb.hooks.OnMessageUpdate_channels, m.ChannelID, func(i int) { sb.hooks.OnMessageUpdate[i].OnMessageUpdate(s, m) })
}
func SBMessageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
  ProcessModules(sb.hooks.OnMessageDelete_channels, m.ChannelID, func(i int) { sb.hooks.OnMessageDelete[i].OnMessageDelete(s, m) })
}
func SBMessageAck(s *discordgo.Session, m *discordgo.MessageAck) { ProcessModules(sb.hooks.OnMessageAck_channels, m.ChannelID, func(i int) { sb.hooks.OnMessageAck[i].OnMessageAck(s, m) }) }
func SBUserUpdate(s *discordgo.Session, u *discordgo.User) {}
func SBPresenceUpdate(s *discordgo.Session, p *discordgo.PresenceUpdate) {}
func SBVoiceStateUpdate(s *discordgo.Session, v *discordgo.VoiceState) {}
func SBGuildUpdate(s *discordgo.Session, g *discordgo.Guild) {
  sb.log.Log("Guild update detected, updating ", g.Name)
  ProcessGuild(g) 
}
func SBGuildMemberAdd(s *discordgo.Session, u *discordgo.Member) { ProcessMember(u) }
func SBGuildMemberRemove(s *discordgo.Session, u *discordgo.Member) { }
func SBGuildMemberDelete(s *discordgo.Session, u *discordgo.Member) { SBGuildMemberRemove(s, u); }
func SBGuildMemberUpdate(s *discordgo.Session, u *discordgo.Member) { ProcessMember(u) }
func SBGuildBanAdd(s *discordgo.Session, b *discordgo.GuildBan) {}
func SBGuildBanRemove(s *discordgo.Session, b *discordgo.GuildBan) {}

func UserHasRole(user uint64, role string) {
  
}

func ProcessUser(u *discordgo.User) uint64 {
  id := SBatoi(u.ID)
  sb.db.AddUser(id, u.Email, u.Username, u.Avatar, u.Verified)
  return id
}

func ProcessMember(u *discordgo.Member) {
  ProcessUser(u.User)
  
  // Parse join date and update user table only if it is less than our current first seen date.
}

func ProcessGuild(g *discordgo.Guild) {
  sb.GuildID = g.ID
  
  for _, v := range g.Channels {
    fmt.Println(v.Name)
    if v.Name == "bot-log" {
      sb.LogChannelID = v.ID
    }
    if v.Name == "ragemuffins" {
      sb.ModChannelID = v.ID
    }
  }
}

func FindChannelID(name string) string {
  channels := sb.dg.State.Guilds[0].Channels 
  for _, v := range channels {
    if v.Name == name {
      return v.ID
    }
  }
  
  return ""
}

func GenChannels(length int, channels *[]map[uint64]bool, fn func(i int) []string) {
  for i := 0; i < length; i++ {
    channel := map[uint64]bool{}
    c := fn(i)
    for j := 0; j < len(c); j++ {
      channel[SBatoi(FindChannelID(c[j]))] = true
    }
    
    *channels = append(*channels, channel)
  }
}

func Initialize() {
  dbauth, _ := ioutil.ReadFile("db.auth")
  discorduser, _ := ioutil.ReadFile("username")  
  discordpass, _ := ioutil.ReadFile("passwd")
  sb = &SweetieBot{}
  sb.version = "0.1.1";
  sb.debug = true
  log := &Log{}
  sb.log = log
  
  db, errdb := DB_Load(log, "mysql", string(dbauth))
  if errdb == nil { defer sb.db.Close(); }
  sb.db = db 
  sb.dg = &discordgo.Session{
    OnEvent: SBEvent,
    OnReady: SBReady,
    OnTypingStart: SBTypingStart,
    OnMessageCreate: SBMessageCreate,
    OnMessageUpdate: SBMessageUpdate,
    OnMessageDelete: SBMessageDelete,
    OnMessageAck: SBMessageAck,
    OnUserUpdate: SBUserUpdate,
    OnPresenceUpdate: SBPresenceUpdate,
    OnVoiceStateUpdate: SBVoiceStateUpdate,
    OnGuildUpdate: SBGuildUpdate,
    OnGuildMemberAdd: SBGuildMemberAdd,
    OnGuildMemberRemove: SBGuildMemberRemove,
    OnGuildMemberUpdate: SBGuildMemberUpdate,
    OnGuildBanAdd: SBGuildBanAdd,
    OnGuildBanRemove: SBGuildBanRemove,
  }
  
  log.Init(sb)
  sb.db.LoadStatements()
  log.Log("Finished loading database statements")  
  log.LogError("Error loading database: ", errdb)
  
  sb.modules = append(sb.modules, &WittyModule{})
  
  for _, v := range sb.modules {
    v.Register(&sb.hooks)
  }
  GenChannels(len(sb.hooks.OnEvent), &sb.hooks.OnEvent_channels, func(i int) []string { return sb.hooks.OnEvent[i].Channels() })
  GenChannels(len(sb.hooks.OnTypingStart), &sb.hooks.OnTypingStart_channels, func(i int) []string { return sb.hooks.OnTypingStart[i].Channels() })
  GenChannels(len(sb.hooks.OnMessageCreate), &sb.hooks.OnMessageCreate_channels, func(i int) []string { return sb.hooks.OnMessageCreate[i].Channels() })
  GenChannels(len(sb.hooks.OnMessageUpdate), &sb.hooks.OnMessageUpdate_channels, func(i int) []string { return sb.hooks.OnMessageUpdate[i].Channels() })
  GenChannels(len(sb.hooks.OnMessageDelete), &sb.hooks.OnMessageDelete_channels, func(i int) []string { return sb.hooks.OnMessageDelete[i].Channels() })
  
  token, err := sb.dg.Login(string(discorduser), string(discordpass))
  if err != nil {
    log.LogError("Discord login failed: ", err)
    return; // this will close the db because we deferred db.Close()
  }
  if token != "" {
      sb.dg.Token = token
  }

  log.LogError("Error opening websocket connection: ", sb.dg.Open());
  log.LogError("Websocket handshake failure: ", sb.dg.Handshake());
  fmt.Println("Connection established");
  log.LogError("Connection error", sb.dg.Listen());
}