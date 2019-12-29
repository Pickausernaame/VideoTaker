package takers

import (
	"fmt"
	"github.com/gocolly/colly"
	"io/ioutil"
	"net/http"
	"strings"
)

func VK_take(link string) (links []string, resolutions []string, err error) {
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
		for _, v := range strs {
			if strings.Contains(v, "RESOLUTION") {
				res_count++
			} else if strings.Contains(v, "https") {
				links = append(links, v)
			}
		}
		resolutions = DEFAULT_RESOLUTIONS[:res_count]
		for i := 0; i < len(links); i++ {
			links[i] = strings.Split(links[i], "/index")[0] + "." + resolutions[i] + ".mp4"
			links[i] = strings.Replace(links[i], "/video/hls", "", 1)
		}
		return
	})

	c.Visit(link)

	return
}
