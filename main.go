package main

import (
	"fmt"
	"github.com/Pickausernaame/VideoTaker/models"
	"github.com/Pickausernaame/VideoTaker/takers/tiktok"
	"github.com/Pickausernaame/VideoTaker/takers/twitter"
	"github.com/Pickausernaame/VideoTaker/takers/vk"
	"github.com/Pickausernaame/VideoTaker/takers/youtube"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func download(uri string, filename string) {
	resp, err := http.Get(uri)
	if err != nil {
		fmt.Println(err)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = ioutil.WriteFile(filename, b, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func main() {

	uri := "https://www.youtube.com/watch?v=pXRviuL6vMY"
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
	case strings.Contains(u.Host, "tiktok.com"):
		vs, err = tiktok.TakeVideo(uri)
		break
	case strings.Contains(u.Host, "youtube.com"), strings.Contains(u.Host, "youtu.be"):
		vs, err = youtube.TakeVideo(uri)
		break
	}

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(vs)
	download(vs[0].Url, "top.mp4")
}
