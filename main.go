package main

import (
	"fmt"
	"github.com/Pickausernaame/VideoTaker/takers"
	"log"
	"strings"
)

const l = "https://vk.com/video?z=video-38316314_456244552%2Fpl_cat_trends"

func ClearLink(link string) string {
	const BASE_URL = "https://m.vk.com/video"
	clear := strings.Split(link, "z=video")
	return BASE_URL + strings.Split(clear[len(clear)-1], "%")[0]
}

func main() {
	link := ClearLink(l)
	fmt.Print(link)
	links, res, err := takers.VK_take(link)
	if err != nil {
		log.Fatal(err)
	}
	for i, v := range links {
		fmt.Println(res[i])
		fmt.Println(v)
	}

}
