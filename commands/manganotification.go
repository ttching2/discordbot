package commands

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/andersfylling/disgord"
	"golang.org/x/net/html"
)

const earlyMangaTimeFormat = "2006-01-02 15:04:05 MST (-07:00)"
const MangaNotificationString = "manga-notification"

type mangaNotificationCommandFactory struct {
	mangaNotificationRepo MangaNotificationRepository
	mangaLinkRepo         MangaLinksRepository
	session               DiscordSession
}

func NewMangaNotificationFactory(repo MangaNotificationRepository, mangaLinkRepo MangaLinksRepository, session DiscordSession) *mangaNotificationCommandFactory {
	return &mangaNotificationCommandFactory{
		mangaNotificationRepo: repo,
		mangaLinkRepo:         mangaLinkRepo,
		session:               session,
	}
}

func (c *mangaNotificationCommandFactory) CreateRequest(data *disgord.MessageCreate, user *Users) interface{} {
	return &mangaNotificationCommand{
		mangaNotificationCommandFactory: c,
		data:                            data,
		user:                            user,
	}
}

type mangaNotificationCommand struct {
	*mangaNotificationCommandFactory
	data *disgord.MessageCreate
	user *Users
}

func (c mangaNotificationCommand) ExecuteMessageCreateCommand() {
	msg := c.data.Message
	split := strings.Split(msg.Content, " ")
	if len(split) != 3 {
		c.session.SendSimpleMessage(msg.ChannelID, "Incorrect number of arguments for command.")
		return
	}
	channelName := split[1]
	guild := c.session.Guild(msg.GuildID)
	channel := FindChannelByName(channelName, guild)
	if channel == nil {
		c.session.SendSimpleMessage(msg.ChannelID, "Channel name not found")
		c.session.ReactWithThumbsDown(msg)
		return
	}

	roleName := split[2]
	roles, _ := c.session.Guild(msg.GuildID).GetRoles()
	role := FindRoleByName(roleName, roles)
	if role == nil {
		c.session.SendSimpleMessage(msg.ChannelID, "Role name not found")
		c.session.ReactWithThumbsDown(msg)
		return
	}

	mangaUrl := split[0]
	mangaLink, err := c.mangaLinkRepo.GetMangaLinkByLink(mangaUrl)
	if err != nil {
		log.Error(err)
		c.session.ReactWithThumbsDown(msg)
		return
	}

	mn := MangaNotification{
		User:     c.user.UsersID,
		Guild:    c.data.Message.GuildID,
		Channel:  channel.ID,
		Role:     role.ID,
	}

	err = c.mangaNotificationRepo.SaveMangaNotification(&mn)

	if err != nil {
		log.Error(err)
		c.session.ReactToMessage(msg.ID, msg.ChannelID, "👎")
		return
	}

	if mangaLink.MangaLinkID == 0 {
		mangaLink = MangaLink{MangaLink: mangaUrl, MangaNotifications: []MangaNotification{mn}}
		c.mangaLinkRepo.SaveMangaLink(&mangaLink)
	}

	err = c.mangaNotificationRepo.AddMangaLink(mn.MangaNotificationID, mangaLink.MangaLinkID)

	if err != nil {
		log.Error(err)
		c.session.ReactToMessage(msg.ID, msg.ChannelID, "👎")
		return
	}
	
	c.session.ReactToMessage(msg.ID, msg.ChannelID, "👍")
}

func LookForNewMangaChapter(repo MangaLinksRepository, s DiscordSession) {
	mangaLinks, err := repo.GetAllMangaLinks()
	if err != nil {
		log.Error(err)
		return
	}
	for _, mangaLink := range mangaLinks {
		go checkEarlyManga(mangaLink, s)
	}
}

func checkEarlyManga(mangaLink MangaLink, s DiscordSession) {
	req, err := http.NewRequest("GET", mangaLink.MangaLink, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "en-US,en;q=0.9")
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
		return
	}
	node, _ := html.Parse(r.Body)
	r.Body.Close()
	if node.FirstChild.NextSibling == nil {
		log.Error("Empty body being fetched from earlymanga")
		return
	}
	c := earlymangacrawler{node, 0, false}
	c.isThereNewChapter(node)
	if c.newChapter {
		for _, guild := range mangaLink.MangaNotifications {
			msg := fmt.Sprintf("%s New chapter found at %s", createMention(guild.Role), mangaLink.MangaLink)
			s.SendSimpleMessage(guild.Channel, msg)
			log.Info("New chapter found at ", mangaLink.MangaLink)
		}
		return
	}
}

type earlymangacrawler struct {
	n          *html.Node
	row        int
	newChapter bool
}

func (t *earlymangacrawler) isThereNewChapter(n *html.Node) {
	body := n.FirstChild.NextSibling.FirstChild.NextSibling.NextSibling
	for child := body.FirstChild; child != nil; child = child.NextSibling {
		if child.Data == "div" {
			for _, attr := range child.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "chapter-row") {
					if t.row == 2 {
						t.row++
						t.newChapter = isChapterNew(child)
						return
					}
					t.row++
				}
			}
		}
		t.isThereNewChapter(child)
	}
}

func isChapterNew(n *html.Node) bool {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		for _, attr := range child.Attr {
			if attr.Key == "title" {
				releaseTime, err := time.Parse(earlyMangaTimeFormat, attr.Val)
				if err != nil {
					log.Error(err)
					return false
				} else {
					now := time.Now().UTC()
					return now.Sub(releaseTime).Hours() <= 1
				}
			}
		}
	}
	return false
}

type mangelocrawler struct {
	n          *html.Node
	newChapter bool
}

func (c *mangelocrawler) findNewChapter(n *html.Node) {
	body := n.FirstChild.FirstChild.NextSibling
	bodySite := body.LastChild.PrevSibling.PrevSibling.PrevSibling.PrevSibling.PrevSibling.PrevSibling.PrevSibling
	containerMain := bodySite.FirstChild.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling
	containerMainLeft := containerMain.FirstChild.NextSibling
	chapterList := containerMainLeft.FirstChild.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling
	chapterRows := chapterList.FirstChild.NextSibling.NextSibling.NextSibling
	firstRow := chapterRows.FirstChild.NextSibling
	time := firstRow.FirstChild.NextSibling.NextSibling.NextSibling.NextSibling.NextSibling
	c.newChapter = c.findTime(time)
}

func (c *mangelocrawler) findTime(n *html.Node) bool {
	for _, attr := range n.Attr {
		if attr.Key == "title" {
			releaseTime, err := time.Parse("Jan 2,2006 15:04", attr.Val)
			if err != nil {
				fmt.Print(err)
			} else {
				now := time.Now().Add(time.Hour * 3)
				return now.Sub(releaseTime).Hours() <= 1
			}
		}
	}

	return false
}
