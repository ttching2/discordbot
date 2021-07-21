package commands

import (
	"bytes"
	"image"
	"image/gif"
	_ "image/gif"
	"image/png"
	"io"
	"regexp"
	"strings"

	"github.com/andersfylling/disgord"
	"github.com/disintegration/gift"
)

const regexpString = "<a?:[a-zA-Z_]+:\\d+>"

type emojifyCommandFactory struct {
	session        DiscordSession
	compiledRegexp *regexp.Regexp
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
	r, _ := regexp.Compile(regexpString)
	return &emojifyCommandFactory{
		session:        s,
		compiledRegexp: r,
	}
}

func (c *emojifyCommand) ExecuteMessageCreateCommand() {
	params := disgord.GetMessagesParams{Before: c.msg.Message.ID, Limit: 1}
	msgs, err := c.session.Channel(c.msg.Message.ChannelID).GetMessages(&params)
	if err != nil {
		log.Error(err)
		c.session.ReactWithThumbsDown(c.msg.Message)
		return
	}
	index := c.compiledRegexp.FindStringIndex(msgs[0].Content)
	if index == nil {
		log.Error("Emoji string not found in content ", msgs[0].Content)
		c.session.ReactWithThumbsDown(c.msg.Message)
		return
	}
	parsedEmojiString := msgs[0].Content[index[0] : index[1]-1]
	firstColon := strings.Index(parsedEmojiString, ":")
	lastColon := strings.LastIndex(parsedEmojiString, ":")
	//Get is animated
	extension := ".png"
	if parsedEmojiString[1:firstColon] == "a" {
		extension = ".gif"
	}
	//Get emoji name
	emojiName := parsedEmojiString[firstColon+1 : lastColon]
	//Get Emoji ID
	emojiId := parsedEmojiString[lastColon+1:]

	response := doHttpGetRequest(discordEmojiCDN + string(emojiId) + extension)

	var emojiFilters []filters
	emojiFilters = append(emojiFilters, resize{2})

	var reader io.Reader
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

func filterPng(src io.Reader, fs []filters) (io.Reader,error) {
	img, _, err := image.Decode(src)
	giftFilter := gift.New()
	for _, f := range fs {
		giftFilter.Add(f.applyFilter(img))
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

func filterGif(src io.Reader, fs []filters) (io.Reader, error) {
	srcGif, err := gif.DecodeAll(src)
	if err != nil {
		return nil, err
	}
	giftFilter := gift.New()
	for _, f := range fs {
		giftFilter.Add(f.applyFilter(srcGif.Image[0]))
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

type filters interface {
	applyFilter(image.Image) gift.Filter
}

type resize struct {
	size int
}

func (f resize) applyFilter(img image.Image) gift.Filter {
	return gift.Resize(img.Bounds().Max.X*f.size, img.Bounds().Max.Y*f.size, gift.BoxResampling)
}