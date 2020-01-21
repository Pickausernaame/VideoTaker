package youtube

import (
	"errors"
	"fmt"
	"github.com/Pickausernaame/VideoTaker/models"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	BASE_URL = "https://www.youtube.com/watch?v="
)

func clearUri(link string) (uri string, err error) {
	id := ""
	switch {
	case strings.Contains(link, "youtube.com"):
		parts := strings.Split(link, "watch?v=")
		if len(parts) < 2 {
			return "", errors.New("Wrong link")
		}
		if strings.Contains(parts[1], "&list=") {
			parts[1] = strings.Split(parts[1], "&list=")[0]
		}
		id = parts[1]
		break
	case strings.Contains(link, "youtu.be"):
		parts := strings.Split(link, "youtu.be/")
		if len(parts) < 2 {
			return "", errors.New("Wrong link")
		}
		id = parts[1]
		break
	default:
		return "", errors.New("Wrong link")
	}
	return BASE_URL + id, nil
}

func downloadVideoStream(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("Non 200 status code received")
	}

	output, _ := ioutil.ReadAll(resp.Body)

	return output, nil
}

type VideoData struct {
	Id         string
	Resolution string
	Type       string
	OnlyVideo  bool
	Link       string
	Size       float32
}

func remove(v VideoDataset, i int) VideoDataset {
	v[i] = v[len(v)-1]
	v = v[:len(v)-1]
	return v
}

type VideoDataset []VideoData

func (a VideoDataset) Len() int           { return len(a) }
func (a VideoDataset) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a VideoDataset) Less(i, j int) bool { return a[i].Size < a[j].Size }

type AudioData struct {
	Id   string
	Type string
	Link string
}

type LinkData struct {
	Id   string
	Link string
}

type Links []LinkData

//todo создавать папки по сессии

func gettingVideoMeta(uri string, wt *sync.WaitGroup, vdata *VideoDataset) {
	defer wt.Done()
	cmd := exec.Command("bash", "getting_meta_av.sh", uri)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	all_videos := strings.Split(string(out), "\n")

	for _, v := range all_videos {

		params := strings.Fields(v)
		if len(params) < 7 {
			continue
		}
		isOnly := strings.Contains(v, "video")
		var size float32
		if strings.Contains(params[len(params)-1], "iB") || strings.Contains(params[len(params)-2], "iB") {

			lenOfPrefix := len(params[len(params)-1])
			k := float32(1024)
			prefix := params[len(params)-1][lenOfPrefix-3 : lenOfPrefix]

			f64, err := strconv.ParseFloat(strings.Split(params[len(params)-1], prefix)[0], 32)

			if err != nil {
				fmt.Println(err)
			}

			size = float32(f64)
			switch prefix {
			case "MiB":
				size = size * k
				break
			case "GiB":
				size = size * k * k
				break
			case "TiB":
				size = size * k * k * k
				break
			}

			d := VideoData{
				Id:         params[0],
				Resolution: strings.Split(params[3], "p")[0],
				Type:       params[1],
				Size:       size,
				OnlyVideo:  isOnly,
			}

			*vdata = append(*vdata, d)
		}

	}
	sort.Sort(vdata)
	for i := len(*vdata) - 1; i > 0; i-- {
		for j := i - 1; j > -1; j-- {
			if ((*vdata)[i].Resolution == (*vdata)[j].Resolution) && ((*vdata)[i].Type == (*vdata)[j].Type) && ((*vdata)[i].OnlyVideo == (*vdata)[j].OnlyVideo) && ((*vdata)[i].Size > (*vdata)[j].Size) {
				*vdata = remove((*vdata), j)
			}

		}
	}
	sort.Sort(vdata)
}

func gettingAudioMeta(uri string, wt *sync.WaitGroup, adata *AudioData) {
	defer wt.Done()

	cmd := exec.Command("bash", "getting_meta_a.sh", uri)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	params := strings.Fields(string(out))
	adata.Id = params[0]
	adata.Type = params[1]
}

func gettingAllLinks(uri string, wt *sync.WaitGroup, links *Links) {
	defer wt.Done()
	cmd := exec.Command("bash", "getting_all_links.sh", uri)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}

	params := strings.Fields(string(out))
	data := LinkData{}
	var re1 = regexp.MustCompile(`(?m)\&(itag=)\d*\&`)
	var re2 = regexp.MustCompile(`(?m)\d+`)
	for _, link := range params {
		data.Id = re2.FindString(re1.FindString(link))
		data.Link = link
		*links = append((*links), data)
	}
}

func TakeVideo(uri string) (videos models.Videos, err error) {
	cluri, _ := clearUri(uri)
	var vdata VideoDataset
	var adata AudioData
	var links Links
	wt := sync.WaitGroup{}
	wt.Add(3)
	gettingVideoMeta(cluri, &wt, &vdata)
	gettingAudioMeta(cluri, &wt, &adata)
	gettingAllLinks(cluri, &wt, &links)
	wt.Wait()
	for _, v := range vdata {
		fmt.Println(v, "\n")
	}
	for _, l := range links {
		for i, _ := range vdata {
			if vdata[i].Id == l.Id {
				vdata[i].Link = l.Link
			}
		}
		if adata.Id == l.Id {
			adata.Link = l.Link
		}
	}

	for _, v := range vdata {
		if !v.OnlyVideo {
			vi := models.Video{
				Url: v.Link,
				Res: v.Resolution,
			}
			videos = append(videos, vi)
		}
	}

	return
}
