package twitter

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Pickausernaame/VideoTaker/models"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

// todo обработать скачивание гифок

const (
	BASE_URL          = "https://twitter.com/i/videos/tweet/"
	ACTIVATE_JSON_URL = "https://api.twitter.com/1.1/guest/activate.json"
	CONFIG_JSON_URL   = "https://api.twitter.com/1.1/videos/tweet/config/"
	AGENT             = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36"
)

func getId(urn string) (id string) {
	re := regexp.MustCompile(`(?m)(.*status\/)|(.*tweet\/)`)
	path := re.FindString(urn)
	id = strings.Replace(urn, path, "", 1)
	return
}

func TakeVideos(urn string) (videos models.Videos, err error) {
	id := getId(urn)
	resp, err := http.Get(BASE_URL + id)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	bearer, err := gettingBearer(resp)
	if err != nil {
		return
	}

	playlist_url := gettingPlaylistUrl(id, bearer, resp)
	if strings.Contains(playlist_url, ".m3u8") {
		videos, err = downloadVideos(playlist_url)
	}
	//} else strings.Contains(playlist_url, ".mp4") {
	//
	//}

	return
}

func gettingBearer(resp *http.Response) (bearer string, err error) {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	// getting js_url from response
	re, _ := regexp.Compile("src=\"(.*)\"")
	js_url := re.FindSubmatch(b)

	if js_url != nil {
		re, _ := regexp.Compile(`(?m)authorization:\"Bearer (.*)\",\"x-csrf`)
		// Get-request to js_url
		resp, err := http.Get(string(js_url[1]))
		if err != nil {
			fmt.Println(err)
			return "", err
		}
		// Converting response to bytes
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return "", err
		}
		// Finding bearer in response
		if r := re.FindSubmatch(b); r != nil {
			bearer = string(r[1])
		}
		resp.Body.Close()
	}
	return
}

func gettingPlaylistUrl(id string, bearer string, resp *http.Response) (playlistUrl string) {
	jsonUrl, _ := url.Parse(CONFIG_JSON_URL + id + ".json")
	USER_AGENT := []string{AGENT}
	ACCEPT_ENCODING := []string{"gzip", "deflate", "br"}
	ORIGIN := []string{"https://twitter.com"}
	X_GUEST_TOKEN := []string{gettingGuestToken(bearer, resp)}
	REFERER := []string{BASE_URL + id}
	AUTHORIZATION := []string{"Bearer " + bearer}

	HEADERS := http.Header{"user-agent": USER_AGENT, "accept-encoding": ACCEPT_ENCODING, "origin": ORIGIN, "x-guest-token": X_GUEST_TOKEN, "referer": REFERER, "authorization": AUTHORIZATION}

	req := &http.Request{
		Method: "GET",
		URL:    jsonUrl,
		Header: HEADERS,
	}

	c := http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	videodataJsonBytes, _ := ioutil.ReadAll(resp.Body)

	t := &track{}
	if err := json.Unmarshal(videodataJsonBytes, &t); err != nil {
		fmt.Println(err)
		return
	}

	clearUrl, _ := url.Parse(t.Track.PlaybackURL)
	playlistUrl = clearUrl.String()
	return
}

func gettingGuestToken(bearer string, resp *http.Response) (token string) {
	url, _ := url.Parse(ACTIVATE_JSON_URL)
	c := http.Client{}

	USER_AGENT := []string{AGENT}
	ACCEPT_ENCODING := []string{"gzip", "deflate", "br"}
	AUTHORIZATION := []string{"Bearer " + bearer}
	COOKIE := gettingCookie(resp)
	HEADERS := http.Header{"user-agent": USER_AGENT, "accept-encoding": ACCEPT_ENCODING, "authorization": AUTHORIZATION, "cookie": COOKIE}

	resp, err := c.Do(&http.Request{URL: url, Method: "POST", Header: HEADERS})
	if err != nil {
		fmt.Println(err)
		return
	}
	token_json_bytes, _ := ioutil.ReadAll(resp.Body)

	var tocken gt
	if err := json.Unmarshal(token_json_bytes, &tocken); err != nil {
		fmt.Println(err)
		return
	}
	return tocken.GuestToken
}

func gettingCookie(resp *http.Response) (cookie []string) {
	var personalization_id, guest_id string
	cookies := resp.Cookies()
	for _, cookie := range cookies {
		switch cookie.Name {
		case "personalization_id":
			personalization_id = cookie.Value
			break
		case "guest_id":
			guest_id = cookie.Value
		}
	}
	cookie = []string{"personalization_id=\"" + personalization_id + "\"; guest_id=" + guest_id}
	return
}

func downloadVideos(url string) (videos models.Videos, err error) {
	switch {
	case strings.Contains(url, ".m3u8"):
		m3u8urls, err := m3u8URLs(url)
		if err != nil {
			fmt.Println(err)
			return videos, err
		}
		c := make(chan models.Video, len(m3u8urls))
		var wg sync.WaitGroup
		wg.Add(len(m3u8urls))

		for _, m3u8 := range m3u8urls {
			go getVideoFromM3U8(m3u8, c, &wg)
		}
		wg.Wait()
		//fmt.Println("All gorutines is complete")

		for i := 0; i < len(m3u8urls); i++ {
			v, ok := <-c
			if !ok {
				break
			}
			videos = append(videos, v)
		}
		break
	}
	return
}

func m3u8URLs(uri string) ([]string, error) {
	if len(uri) == 0 {
		return nil, errors.New("url is null")
	}
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	html := string(bytes)
	//html, err := request.Get(uri, "", nil)
	//if err != nil {
	//	return nil, err
	//}
	lines := strings.Split(html, "\n")
	var urls []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			if strings.HasPrefix(line, "http") {
				urls = append(urls, line)
			} else {
				base, err := url.Parse(uri)
				if err != nil {
					continue
				}
				u, err := url.Parse(line)
				if err != nil {
					continue
				}
				urls = append(urls, fmt.Sprintf("%s", base.ResolveReference(u)))
			}
		}
	}
	return urls, nil
}

func getVideoFromM3U8(m3u8 string, c chan models.Video, wg *sync.WaitGroup) {
	defer wg.Done()
	fileName := extractFilename(m3u8)
	dir := "./twitter/"
	fileName = dir + strings.Split(fileName, ".")[0] + ".ts"
	ts, err := m3u8URLs(m3u8)
	if err != nil {
		fmt.Println(err)
		return
	}
	var files []string
	for _, v := range ts {
		file, _ := getVideoPart(v)
		files = append(files, file)
	}
	combineTs(fileName, files)

	re := regexp.MustCompile(`(?m)\d*x\d*`)
	resolution := re.FindAllString(m3u8, 2)
	mp4 := models.Video{
		Url: convertTStoMP4(fileName),
		Res: resolution[1],
	}
	c <- mp4
	return
}

func extractFilename(url string) string {
	strs := strings.Split(url, "/")
	fileName := strs[len(strs)-1]

	params := strings.Split(fileName, "&")
	param := params[len(params)-1]
	return param
}

func getVideoPart(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Encoding", "gzip,deflate,br")

	c := &http.Client{}
	resp, err := c.Do(req)

	defer resp.Body.Close()
	if err != nil {
		return "", err
	}
	fileName := "./twitter/" + extractFilename(url)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(fileName, body, os.ModePerm)
	if err != nil {
		return "", err
	}
	return fileName, nil
}

func combineTs(filename string, files []string) {
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		return
	}
	writeLen := 0
	for _, v := range files {
		data, err := ioutil.ReadFile(v)
		if err != nil {
			return
		}
		file.WriteAt(data, int64(writeLen))
		writeLen += len(data)
		os.Remove(v)
	}

}

func convertTStoMP4(tsFilename string) (mp4Filename string) {
	resultDir := "./mp4/"
	mp4Filename = resultDir + strings.Replace(strings.Replace(tsFilename, ".ts", ".mp4", 1), "./twitter/", "", 1)
	convert := exec.Command("ffmpeg", "-i", tsFilename, "-c", "copy", mp4Filename, "-y")

	//convert.Stdout = os.Stdout
	//convert.Stderr = os.Stderr
	convert.Run()
	os.Remove(tsFilename)
	return
}
