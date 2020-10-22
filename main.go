package main

import (
	"fmt"
	"strings"
	"database/sql"
	"github.com/gocolly/colly"
	_ "github.com/lib/pq"
)

var hangul = []string{"ㄱ", "ㄲ", "ㄴ", "ㄷ", "ㄸ", "ㄹ", "ㅁ", "ㅂ", "ㅃ", "ㅅ", "ㅆ", "ㅇ", "ㅈ", "ㅉ", "ㅊ", "ㅋ", "ㅌ", "ㅍ", "ㅎ"}
var songIDList = []int{
	32821268, 32720312, 32550258, 32821269, 32074793, 31558845, 31556642, 31360057, 30945748, 30945750, 30442305, 30183395, 30442306,
	32961283, 32442735, 32224085, 31978150, 32091615, 31815490, 31626257, 31521709, 31219546, 31554317, 30773554, 32720013, 32825454,
	32438894, 32224272, 32734372, 32987478, 32999767, 32646938, 30613202, 32224271, 5835766,  5620266, 3843566, 1637914, 5777833, 5674119,
	1556553, 31340985, 30433125, 32832197, 32643130, 32643135, 32616844, 32578369, 32411219, 32224280, 32559498, 32079740, 31796505, 31284909,
	31123622, 8102118, 8102122,
}
var searchKeyWord = "ㄴ ㄴㄴㄴㄴ ㄴㄴㄴㄴㄴ ㄴㄴ ㄴㄴ ㄴㄴㄴ ㄴㄴ ㄴㄴ"

type songInfo struct {
	id		int
	title	string
	artist	string
	lyrics	string
}

func getSongInfo(songID int, db *sql.DB, mainScrapCh chan<- songInfo) {
	var s songInfo
	s.id = songID

	query := fmt.Sprintf("SELECT COUNT(*) AS cnt FROM melon WHERE id='%d'", s.id)
	rows, err := db.Query(query)
	checkErr(err)

	if countRows(rows) == 1 {
		query2 := fmt.Sprintf("SELECT title, artist, lyrics FROM melon WHERE id='%d'", s.id)
		rows2, err2 := db.Query(query2)
		checkErr(err2)

		for rows2.Next() {
			rows2.Scan(&s.title, &s.artist, &s.lyrics)
		}

		mainScrapCh <- s
	} else {
		c := colly.NewCollector(
			colly.AllowedDomains("www.melon.com"),
		)
	
		c.OnHTML("div", func(e *colly.HTMLElement) {
			id := e.Attr("id")
			class := e.Attr("class")
	
			if class == "song_name" {
				s.title = e.Text
				s.title = fmt.Sprintf("%q", s.title)
				s.title = strings.ReplaceAll(s.title, "\"", "")
				s.title = strings.ReplaceAll(s.title, "\\t", "")
				s.title = strings.ReplaceAll(s.title, "\\n", "")
				s.title = strings.ReplaceAll(s.title, "&#39;", "'")
				s.title = strings.ReplaceAll(s.title, "곡명", "")
	
				c.Visit(e.Request.AbsoluteURL(class))
			}
	
			if class == "artist" {
				s.artist = e.DOM.Children().Text()
			}
			
			if id == "d_video_summary" {
				temp, _ := e.DOM.Html()
				temp = fmt.Sprintf("%q", temp)
				temp = strings.ReplaceAll(temp, "<!-- height:auto; 로 변경시, 확장됨 -->", "")
				temp = strings.ReplaceAll(temp, "\"", "")
				temp = strings.ReplaceAll(temp, "\\t", "")
				temp = strings.ReplaceAll(temp, "\\n", "")
				temp = strings.ReplaceAll(temp, "&#39;", "'")
				temp = strings.ReplaceAll(temp, "<br/>", " ")
				
				for _, ascii := range(temp) {	
					if 44032 <= ascii && ascii <= 55203 {
						s.lyrics += hangul[(ascii - 44032) / 588]
					} else if ascii == 32 {
						s.lyrics += " "
					}
				}
	
				for ;strings.Count(s.lyrics, "  ") > 0; {
					s.lyrics = strings.ReplaceAll(s.lyrics, "  ", " ")
				}
	
				c.Visit(e.Request.AbsoluteURL(id))
			}
		})
	
		c.Visit(fmt.Sprintf("https://www.melon.com/song/detail.htm?songId=%d", songID))

		query2 := fmt.Sprintf("INSERT INTO melon(id, title, artist, lyrics) VALUES('%d', '%s', '%s', '%s')", s.id, s.title, s.artist, s.lyrics)
		_, err2 := db.Query(query2)
		checkErr(err2)
	
		mainScrapCh <- s
	}
}

func searchKeyword(s songInfo, mainSearchCh chan<- bool) {
	res := false

	if strings.Count(s.lyrics, searchKeyWord) > 0 {
		res = true
	}

	mainSearchCh <- res 
}

func main() {
	db := connectToDB("goInitialLyrics")
	checkErr(db.Ping())
	defer db.Close()
	
	mainScrapCh := make(chan songInfo)

	for _, id := range(songIDList) {
		go getSongInfo(id, db, mainScrapCh)
	}

	mainSearchCh := make(chan bool)
	var songList []songInfo

	for i := 0; i < len(songIDList); i++ {
		s := <-mainScrapCh
		songList = append(songList, s)
		go searchKeyword(s, mainSearchCh)
	}

	var isExistsList []bool

	cnt := 0

	for i := 0; i < len(songIDList); i++ {
		isExists := <-mainSearchCh
		isExistsList = append(isExistsList, isExists)

		if isExists {
			fmt.Println(songList[i].title)
			cnt++
		}
	}

	if cnt == 0 {
		fmt.Println("결과를 못 찾았어요")
	}
}

func connectToDB(dbname string) *sql.DB {
	db, err := sql.Open("postgres",
	fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
	"localhost", 5432, "postgres", "패스워드", dbname))
	checkErr(err)
  
	return db
}

func countRows(rows *sql.Rows) (count int) {
	for rows.Next() {
    	err := rows.Scan(&count)
    	checkErr(err)
  	}

  	return count
}

func checkErr(err error) {
	if err != nil {
	  panic(err)
	}
}