package main

import (
	"fmt"
	"strings"
)

const l = "https://vk.com/video?z=video-36775085_456240117%2Fpl_cat_63"

func ClearLink(link string) string {
	const BASE_URL = "https://m.vk.com/video"
	clear := strings.Split(link, "z=video")
	return BASE_URL + strings.Split(clear[len(clear)-1], "%")[0]
}

func main() {
	link := ClearLink(l)
	fmt.Print(link)

}
