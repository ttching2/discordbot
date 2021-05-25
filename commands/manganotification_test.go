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

func TestCreateMangaNotification(t *testing.T) {
	const content = "https://earlymanga.org/manga/a-returner-s-magic-should-be-special channel role"
	repo := mockMangaNotificationRepo{}
	s := &mockSession{
		guild: &commonMockGuild,
	}
	msg := &disgord.MessageCreate{
		Message: &disgord.Message{
			Content: content,
		},
	}
	user := &commands.Users{UsersID: 1, DiscordUsersID: 1}
	factory := commands.NewMangaNotificationFactory(&repo, s)
	c := factory.CreateRequest(msg, user)
	c.(onMessageCreateCommand).ExecuteMessageCreateCommand()
	result, _ := repo.GetAllMangaNotifications()

	if len(result) != 1 {
		log.Println("")
		t.Fail()
	}
}