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
		User:     c.user.UsersID,
		Guild:    c.data.Message.GuildID,
		MangaURL: split[0],
		Channel:  channel.ID,
		Role:     role.ID,
	}

	err := c.repo.SaveMangaNotification(&mn)

	if err != nil {
		log.Error(err)
		c.session.ReactToMessage(msg.ID, msg.ChannelID, "👎")
		return
	}
	c.session.ReactToMessage(msg.ID, msg.ChannelID, "👍")
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
	log.Info("Starting search for manga chapter in ", mangaNotification.MangaURL)
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
	c := earlymangacrawler{body, 0, false}
	c.isThereNewChapter(body)
	if  c.newChapter {
		msg := fmt.Sprintf("%s New chapter found at %s", createMention(mangaNotification.Role), mangaNotification.MangaURL)
		s.SendSimpleMessage(mangaNotification.Channel, msg)
		log.Info("New chapter found at ", mangaNotification.MangaURL)
		return
	}
	log.Info("No new chapter found at ", mangaNotification.MangaURL)
}

type earlymangacrawler struct {
	n          *html.Node
	row        int
	newChapter bool
}

func (t *earlymangacrawler) isThereNewChapter(n *html.Node) {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
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
