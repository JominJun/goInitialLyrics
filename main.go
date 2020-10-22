package main

import (
	"fmt"
	"strings"
	"github.com/fatih/color"
	"github.com/gocolly/colly"
)

var hangul = []string{"ㄱ", "ㄲ", "ㄴ", "ㄷ", "ㄸ", "ㄹ", "ㅁ", "ㅂ", "ㅃ", "ㅅ", "ㅆ", "ㅇ", "ㅈ", "ㅉ", "ㅊ", "ㅋ", "ㅌ", "ㅍ", "ㅎ"}
var songIDList = []int{
	32821268, 32720312, 32550258, 32821269, 32074793, 31558845, 31556642, 31360057, 30945748, 30945750, 30442305, 30183395, 30442306,
	32961283, 32442735, 32224085, 31978150, 32091615, 31815490, 31626257, 31521709, 31219546, 31554317, 30773554, 32720013, 32825454,
	32438894, 32224272, 32734372, 32987478, 32999767, 32646938, 30613202, 32224271, 5835766,  5620266, 3843566, 1637914, 5777833, 5674119,
	1556553, 31340985, 30433125,
}
var searchKeyWord = "ㅈㅅㄹㅇ ㄴㄷㄱ"

func main() {
	for _, songID := range(songIDList) {
		var songName = ""
		var artist = ""
		var purifiedLyrics = ""

		c := colly.NewCollector(
			colly.AllowedDomains("www.melon.com"),
		)

		c.OnHTML("div", func(e *colly.HTMLElement) {
			id := e.Attr("id")
			class := e.Attr("class")

			if class == "song_name" {
				songName = e.Text
				songName = fmt.Sprintf("%q", songName)
				songName = strings.ReplaceAll(songName, "\"", "")
				songName = strings.ReplaceAll(songName, "\\t", "")
				songName = strings.ReplaceAll(songName, "\\n", "")
				songName = strings.ReplaceAll(songName, "곡명", "")

				c.Visit(e.Request.AbsoluteURL(class))
			}

			if class == "artist" {
				artist = e.DOM.Children().Text()
			}
			
			if id == "d_video_summary" {
				lyrics, _ := e.DOM.Html()
				lyrics = fmt.Sprintf("%q", lyrics)
				lyrics = strings.ReplaceAll(lyrics, "<!-- height:auto; 로 변경시, 확장됨 -->", "")
				lyrics = strings.ReplaceAll(lyrics, "\"", "")
				lyrics = strings.ReplaceAll(lyrics, "\\t", "")
				lyrics = strings.ReplaceAll(lyrics, "\\n", "")
				lyrics = strings.ReplaceAll(lyrics, "&#39;", "'")
				lyrics = strings.ReplaceAll(lyrics, "<br/>", " ")
				
				for _, ascii := range(lyrics) {	
					if 44032 <= ascii && ascii <= 55203 {
						purifiedLyrics += hangul[(ascii - 44032) / 588]
					} else if ascii == 32 {
						purifiedLyrics += " "
					}
				}

				for ;strings.Count(purifiedLyrics, "  ") > 0; {
					purifiedLyrics = strings.ReplaceAll(purifiedLyrics, "  ", " ")
				}

				c.Visit(e.Request.AbsoluteURL(id))
			}
		})

		c.OnRequest(func(r *colly.Request) {
			//fmt.Println("Visiting", r.URL.String())
		})

		c.Visit(fmt.Sprintf("https://www.melon.com/song/detail.htm?songId=%d", songID))

		if strings.Count(purifiedLyrics, searchKeyWord) > 0 {
			color.Green(songName)
			color.Blue(artist)
		} else {
			color.Red(songName)
			color.Blue(artist)
		}
		
		//fmt.Println(strings.TrimSpace(purifiedLyrics))
		fmt.Println()
	}
}