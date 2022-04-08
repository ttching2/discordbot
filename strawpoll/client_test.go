package strawpoll_test

import (
	"discordbot/strawpoll"
	"fmt"
	"image"
	"os"
	"testing"
	"time"

	"golang.org/x/image/draw"
)

func thingy(v string) {
	fmt.Println(v)
}
//Channel 803072579765403669
func TestThing(t *testing.T) {
	client := strawpoll.New(strawpoll.StrawPollConfig{
		ApiKey: os.Getenv("STRAWPOLL_TOKEN"),
	})
	r, _ := client.GetPoll("05Zd1mAaEy6")
	thing := time.Unix(r.Poll.PollConfig.DeadlineAt, 0)
	fmt.Println(thing)
	// client := disgord.New(disgord.Config{
	// 	BotToken: os.Getenv("DISCORD_TOKEN"),
	// })
	// f1, err := os.Open("imgs.jpg")
	// if err != nil {
	// 	panic(err)
	// }
	// _, errUpload := client.WithContext(context.Background()).SendMsg(803072579765403669, &disgord.CreateMessageParams{
	// 	Content: "",
	// 	Files: []disgord.CreateMessageFileParams{
	// 		{f1, "myfavouriteimage.jpg", false},
	// 	},
	// })
	// if errUpload != nil {
	// 	client.Logger().Error("unable to upload images.", errUpload)
	// }
	// req, err := http.NewRequest("GET", "https://cdn.discordapp.com/emojis/776012671936888843.png", nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// r, err := http.DefaultClient.Do(req)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer r.Body.Close()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// img,_, err := image.Decode(r.Body)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// out, _ := os.Create("./img.png")
    // defer out.Close()
	// filter := gift.New(
	// 	gift.Resize(img.Bounds().Max.X*2, img.Bounds().Max.Y*2, gift.BoxResampling),
	// )
	// dr := image.NewRGBA(filter.Bounds(img.Bounds()))
	// filter.Draw(dr, img)
	// err = png.Encode(out,dr)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}

// src   - source image
// rect  - size we want
// scale - scaler
func scaleTo(src image.Image, rect image.Rectangle, scale draw.Scaler) image.Image {
	dst := image.NewRGBA(rect)
	scale.Scale(dst, rect, src, src.Bounds(), draw.Over, nil)
	return dst
}