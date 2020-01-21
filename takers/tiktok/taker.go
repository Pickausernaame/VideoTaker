package tiktok

import (
	"github.com/Pickausernaame/VideoTaker/models"
	"github.com/gocolly/colly"
)

func TakeVideo(uri string) (videos models.Videos, err error) {
	c := colly.NewCollector()
	c.OnHTML("video", func(e *colly.HTMLElement) {

		v := models.Video{}
		v.Url = e.Attr("src")
		videos = append(videos, v)

	})
	c.Visit(uri)
	return
}
