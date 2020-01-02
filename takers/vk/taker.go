package vk

import (
	"fmt"
	"github.com/Pickausernaame/VideoTaker/models"
	"github.com/gocolly/colly"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	BASE_URL = "https://m.vk.com/video"
)

func clearUri(link string) string {
	clear := strings.Split(link, "z=video")
	return BASE_URL + strings.Split(clear[len(clear)-1], "%")[0]
}

func TakeVideo(uri string) (videos models.Videos, err error) {
	cluri := clearUri(uri)
	c := colly.NewCollector()
	DEFAULT_RESOLUTIONS := [5]string{"240", "360", "480", "720", "1080"}

	c.OnHTML("source[type='application/vnd.apple.mpegurl']", func(e *colly.HTMLElement) {
		resp, err := http.Get(e.Attr("src"))
		if err != nil {
			fmt.Println(err)
			return
		}
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		strs := strings.Split(string(b), "\n")
		res_count := 0
		var links []string
		var resolutions []string
		for _, v := range strs {
			if strings.Contains(v, "RESOLUTION") {
				res_count++
			} else if strings.Contains(v, "https") {
				links = append(links, v)
			}
		}
		resolutions = DEFAULT_RESOLUTIONS[:res_count]
		v := models.Video{}
		for i := 0; i < len(links); i++ {

			links[i] = strings.Split(links[i], "/index")[0] + "." + resolutions[i] + ".mp4"
			links[i] = strings.Replace(links[i], "/video/hls", "", 1)
			v.Res = resolutions[i]
			v.Url = links[i]
			videos = append(videos, v)
		}
		return
	})

	c.Visit(cluri)

	return
}
