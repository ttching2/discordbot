package commands_test

import (
	"discordbot/commands"
	"log"
	"testing"

	"github.com/andersfylling/disgord"
)

type mockMangaNotificationRepo struct {
	id int64
	notifications []commands.MangaNotification
	links []xref
}

type xref struct {
	notifId int64
	linkId int64
}

type mockMangaLinkRepo struct {
	id int64
	link []commands.MangaLink
}

func (r *mockMangaNotificationRepo) SaveMangaNotification(c *commands.MangaNotification) error {
	c.MangaNotificationID = r.id
	r.id++
	r.notifications = append(r.notifications, *c)
	return nil
}

func (r *mockMangaNotificationRepo) GetAllMangaNotifications() ([]commands.MangaNotification, error) {
	return r.notifications, nil
}
func (r *mockMangaNotificationRepo) AddMangaLink(mangaNotificationId int64, mangaLinkId int64) error {
	r.links = append(r.links, xref{mangaNotificationId, mangaLinkId})
	return nil
}

func (r *mockMangaLinkRepo) SaveMangaLink(link *commands.MangaLink) error {
	r.link = append(r.link, *link)
	return nil
}

func (r *mockMangaLinkRepo) GetMangaLinkByLink( link string) (commands.MangaLink, error) {
	for _, l := range r.link {
		if l.MangaLink == link {
			return l, nil
		}
	}
	return commands.MangaLink{}, nil
}

func (r *mockMangaLinkRepo) GetAllMangaLinks() ([]commands.MangaLink, error) {
	return r.link, nil
}

func TestCreateMangaNotification(t *testing.T) {
	const content = "https://earlymanga.org/manga/a-returner-s-magic-should-be-special channel role"
	repo := mockMangaNotificationRepo{}
	lrepo := mockMangaLinkRepo{}
	s := &mockSession{
		guild: &commonMockGuild,
	}
	msg := &disgord.MessageCreate{
		Message: &disgord.Message{
			Content: content,
		},
	}
	user := &commands.Users{UsersID: 1, DiscordUsersID: 1}
	factory := commands.NewMangaNotificationFactory(&repo, &lrepo, s)
	c := factory.CreateRequest(msg, user)
	c.(onMessageCreateCommand).ExecuteMessageCreateCommand()
	result, _ := repo.GetAllMangaNotifications()
	linkResult, _ := lrepo.GetAllMangaLinks()

	if len(result) != 1 {
		log.Println("Manga Notification not saved")
		t.FailNow()
	}
	if len(linkResult) != 1 {
		log.Println("Manga Link not saved")
		t.FailNow()
	}
}