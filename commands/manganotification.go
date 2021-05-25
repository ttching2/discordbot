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

type mangaNotificationCommandFactory struct {
	repo    MangaNotificationRepository
	session DiscordSession
}

func NewMangaNotificationFactory(repo MangaNotificationRepository, session DiscordSession) *mangaNotificationCommandFactory {
	return &mangaNotificationCommandFactory{
		repo:    repo,
		session: session,
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

	roleName := split[2]
	roles, _ := c.session.Guild(msg.GuildID).GetRoles()
	role := FindRoleByName(roleName, roles)

	mn := MangaNotification{
		MangaURL: split[0],
		Channel:  channel.ID,
		Role:     role.ID,
	}

	err := c.repo.SaveMangaNotification(&mn)

	if err != nil {
		log.Error(err)
	}
}

func LookForNewMangaChapter(repo MangaNotificationRepository, s DiscordSession) {
	mangaLinks, err := repo.GetAllMangaNotifications()
	if err != nil {
		log.Error(err)
		return
	}
	for _, mangaLink := range mangaLinks {
		go checkEarlyManga(mangaLink, s)
	}
}

func checkEarlyManga(mangaNotification MangaNotification, s DiscordSession) {
	req, err := http.NewRequest("GET", mangaNotification.MangaURL, nil)
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
	body := node.FirstChild.NextSibling.FirstChild.NextSibling.NextSibling
	c := earlymangacrawler{body, 0}
	if c.isThereNewChapter(body) {
		msg := fmt.Sprintf("%s New chapter found at %s", createMention(mangaNotification.Role), mangaNotification.MangaURL)
		s.SendSimpleMessage(mangaNotification.Channel, msg)
	}
}

type earlymangacrawler struct {
	n   *html.Node
	row int
}

func (t *earlymangacrawler) isThereNewChapter(n *html.Node) bool {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Data == "div" {
			for _, attr := range child.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "chapter-row") {
					if t.row == 2 {
						return isChapterNew(child)
					}
					t.row++
				}
			}
		}
		t.isThereNewChapter(child)
	}
	return false
}

func isChapterNew(n *html.Node) bool {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		for _, attr := range child.Attr {
			if attr.Key == "title" {
				time, err := time.Parse(earlyMangaTimeFormat, attr.Val)
				if err != nil {
					log.Error(err)
					return false
				} else {
					now := time.Local().UTC()
					return now.Sub(time).Hours() <= 1
				}
			}
		}
	}
	return false
}
