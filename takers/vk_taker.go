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
		for _, v := range strs {
			if strings.Contains(v, "RESOLUTION") {
				res := strings.Split(v, "x")
				resolutions = append(resolutions, res[len(res)-1])
			} else if strings.Contains(v, "https") {
				links = append(links, v)
			}
		}
		for i := 0; i < len(links); i++ {
			links[i] = strings.Split(links[i], "/index")[0] + "." + resolutions[i] + ".mp4"
			links[i] = strings.Replace(links[i], "/video/hls", "", 1)
		}
		for _, v := range links {
			fmt.Println(v)
		}
	})
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})
	c.Visit(link)
	return
}
