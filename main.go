package main

import (
	"fmt"
	"github.com/Pickausernaame/VideoTaker/models"
	"github.com/Pickausernaame/VideoTaker/takers/twitter"
	"github.com/Pickausernaame/VideoTaker/takers/vk"
	"net/url"
	"strings"
)

func main() {

	uri := "https://vk.com/video?z=video-45745333_456276145%2Fpl_cat_trends"
	var vs models.Videos
	u, err := url.Parse(uri)
	if err != nil {
		fmt.Println(err)
		return
	}
	switch {
	case strings.Contains(u.Host, "vk.com"):
		vs, err = vk.TakeVideo(uri)
		break
	case strings.Contains(u.Host, "twitter.com"):
		vs, err = twitter.TakeVideos(uri)
		break
	}

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(vs)

}
