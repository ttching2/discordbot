package commands

import (
	"bytes"
	"image"
	"image/gif"
	_ "image/gif"
	"image/png"
	"io"
	"math/rand"
	"regexp"
	"strings"

	"github.com/andersfylling/disgord"
	"github.com/disintegration/gift"
)

const discordEmojiFormat = "<a?:[0-9a-zA-Z_]+:\\d+>"

type emojifyCommandFactory struct {
	session     DiscordSession
	emojiParser *regexp.Regexp
}

type emojifyCommand struct {
	*emojifyCommandFactory
	msg  *disgord.MessageCreate
	user *Users
}

func (c *emojifyCommandFactory) CreateRequest(data *disgord.MessageCreate, user *Users) interface{} {
	return &emojifyCommand{
		c,
		data,
		user,
	}
}

func NewEmojifyCommandFactory(s DiscordSession) *emojifyCommandFactory {
	r, _ := regexp.Compile(discordEmojiFormat)
	return &emojifyCommandFactory{
		session:     s,
		emojiParser: r,
	}
}

func (c *emojifyCommand) ExecuteMessageCreateCommand() {
	var emojiString, emoteArgs string
	emojiIndex := c.emojiParser.FindStringIndex(c.msg.Message.Content)
	if emojiIndex != nil {
		emojiString = c.msg.Message.Content[emojiIndex[0] : emojiIndex[1]-1]
		emoteArgs = c.msg.Message.Content[emojiIndex[1]:]
		
	} else {
		params := disgord.GetMessagesParams{Before: c.msg.Message.ID, Limit: 1}
		msgs, err := c.session.Channel(c.msg.Message.ChannelID).GetMessages(&params)
		
		if err != nil {
			log.Error(err)
			c.session.ReactWithThumbsDown(c.msg.Message)
			return
		}
		index := c.emojiParser.FindStringIndex(msgs[0].Content)
		if index == nil {
			log.Error("Emoji string not found in content ", msgs[0].Content)
			c.session.ReactWithThumbsDown(c.msg.Message)
			return
		}
		emojiString = msgs[0].Content[index[0] : index[1]-1]
		emoteArgs = c.msg.Message.Content
	}
	
	firstColon := strings.Index(emojiString, ":")
	lastColon := strings.LastIndex(emojiString, ":")
	//Get is animated
	extension := ".png"
	if emojiString[1:firstColon] == "a" {
		extension = ".gif"
	}
	//Get emoji name
	emojiName := emojiString[firstColon+1 : lastColon]
	//Get Emoji ID
	emojiId := emojiString[lastColon+1:]

	response := doHttpGetRequest(discordEmojiCDN + string(emojiId) + extension)
	var emojiFilters []func(image.Image) gift.Filter
	if len(emoteArgs) > 0 {
		emojiFilters = append(emojiFilters, parseArguments(emoteArgs)...)
	}
	emojiFilters = append(emojiFilters,
		func(img image.Image) gift.Filter {
			return gift.Resize(img.Bounds().Max.X*2, img.Bounds().Max.Y*2, gift.BoxResampling)
		})

	var reader io.Reader
	var err error
	if extension == ".png" {
		reader, err = filterPng(response, emojiFilters)
	} else if extension == ".gif" {
		reader, err = filterGif(response, emojiFilters)
	}
	if err != nil {
		log.Error(err)
		c.session.ReactWithThumbsDown(c.msg.Message)
	}
	fileMsg := disgord.CreateMessageFileParams{
		Reader:     reader,
		FileName:   emojiName + extension,
		SpoilerTag: false,
	}
	_, err = c.session.SendMessage(c.msg.Message.ChannelID, &disgord.CreateMessageParams{Files: []disgord.CreateMessageFileParams{fileMsg}})
	if err != nil {
		log.Error(err)
		c.session.ReactWithThumbsDown(c.msg.Message)
		return
	}
}

func parseArguments(args string) []func(image.Image) gift.Filter {
	var fs []func(image.Image) gift.Filter
	for _, s := range strings.ToLower(args) {
		switch s {
		case 'p':
			f := func(img image.Image) gift.Filter { return gift.Pixelate(5) }
			fs = append(fs, f)
		case 'i':
			f := func(img image.Image) gift.Filter { return gift.Invert() }
			fs = append(fs, f)
		case 'r':
			f := func(img image.Image) gift.Filter { return gift.Rotate90() }
			fs = append(fs, f)
		case 'b':
			f := func(img image.Image) gift.Filter { return gift.GaussianBlur(3) }
			fs = append(fs, f)
		case 'h':
			f := func(img image.Image) gift.Filter { return gift.Hue(90) }
			fs = append(fs, f)
		case 'c':
			f := func(img image.Image) gift.Filter {
				return gift.CropToSize(rand.Intn(img.Bounds().Max.X), rand.Intn(img.Bounds().Max.Y), randomAnchor())
			}
			fs = append(fs, f)
		}
	}
	return fs
}

func randomAnchor() gift.Anchor {
	switch rand.Intn(9) {
	case 0:
		return gift.CenterAnchor
	case 1:
		return gift.TopLeftAnchor
	case 2:
		return gift.TopAnchor
	case 3:
		return gift.TopRightAnchor
	case 4:
		return gift.LeftAnchor
	case 5:
		return gift.RightAnchor
	case 6:
		return gift.BottomLeftAnchor
	case 7:
		return gift.BottomAnchor
	case 8:
		return gift.BottomRightAnchor
	}
	return gift.CenterAnchor
}

func filterPng(src io.Reader, fs []func(image.Image) gift.Filter) (io.Reader, error) {
	img, _, err := image.Decode(src)
	giftFilter := gift.New()
	for _, f := range fs {
		giftFilter.Add(f(img))
	}
	destinationImg := image.NewRGBA(giftFilter.Bounds(img.Bounds()))
	giftFilter.Draw(destinationImg, img)
	if err != nil {
		return nil, err
	}
	buff := new(bytes.Buffer)
	err = png.Encode(buff, destinationImg)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buff.Bytes()), nil
}

func filterGif(src io.Reader, fs []func(image.Image) gift.Filter) (io.Reader, error) {
	srcGif, err := gif.DecodeAll(src)
	if err != nil {
		return nil, err
	}
	giftFilter := gift.New()
	for _, f := range fs {
		giftFilter.Add(f(srcGif.Image[0]))
	}
	srcGif.Config.Height = giftFilter.Bounds(srcGif.Image[0].Rect).Max.Y
	srcGif.Config.Width = giftFilter.Bounds(srcGif.Image[0].Rect).Max.X
	for i, _ := range srcGif.Image {
		frame := srcGif.Image[i]
		dst := image.NewPaletted(giftFilter.Bounds(frame.Bounds()), frame.Palette)
		giftFilter.Draw(dst, frame)
		srcGif.Image[i] = dst
	}
	buff := new(bytes.Buffer)
	err = gif.EncodeAll(buff, srcGif)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buff.Bytes()), nil
}
