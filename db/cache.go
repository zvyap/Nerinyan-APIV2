package db

import (
	"github.com/Nerinyan/Nerinyan-APIV2/osu"
	"github.com/dchest/stemmer/porter2"
	"regexp"
	"strings"
)

var (
	//STRING_INDEX     = map[string]*struct{}{}
	regexpReplace, _ = regexp.Compile(`[^0-9A-z]|[\[\]]`)
)

func LoadCache() {
	//rows, err := Maria.Query(`select STRING from SEARCH_CACHE_STRING_INDEX`)
	//if err != nil && err != sql.ErrNoRows {
	//	pterm.Error.Println(err)
	//	return
	//}
	//defer rows.Close()
	//var tmp string
	//for rows.Next() {
	//	err := rows.Scan(&tmp)
	//	if err != nil {
	//		pterm.Error.Println(err)
	//		continue
	//	}
	//	STRING_INDEX[tmp] = &struct{}{}
	//}
	//pterm.Info.Println("LoadCache() end")
}

func InsertCache(data *[]osu.BeatmapSetsIN) {
	go insertStringIndex(data)

}

type insertData struct {
	Strbuf  []string
	Artist  []row
	Creator []row
	Title   []row
	Tags    []row
}
type row struct {
	KEY          []string
	BeatmapsetId int
}

func insertStringIndex(data *[]osu.BeatmapSetsIN) {
	//defer func() {
	//	err, e := recover().(error)
	//	if e {
	//		pterm.Error.Println(err)
	//	}
	//}()
	insertData := insertData{}

	for _, in := range *data {
		artist := splitString(*in.Artist)
		creator := splitString(*in.Creator)
		title := splitString(*in.Title)
		tags := splitString(*in.Tags)
		insertData.Artist = append(insertData.Artist, row{
			KEY:          artist,
			BeatmapsetId: in.Id,
		})
		insertData.Creator = append(insertData.Creator, row{
			KEY:          creator,
			BeatmapsetId: in.Id,
		})
		insertData.Title = append(insertData.Title, row{
			KEY:          title,
			BeatmapsetId: in.Id,
		})
		insertData.Tags = append(insertData.Tags, row{
			KEY:          tags,
			BeatmapsetId: in.Id,
		})
		insertData.Strbuf = append(insertData.Strbuf, artist...)
		insertData.Strbuf = append(insertData.Strbuf, creator...)
		insertData.Strbuf = append(insertData.Strbuf, title...)
		insertData.Strbuf = append(insertData.Strbuf, tags...)

	}

	//pterm.Println(string(*utils.ToJsonString(makeArrayUnique(insertData.Strbuf))))
	//panic("")
	err := BulkInsertLimiter(
		"INSERT IGNORE INTO SEARCH_CACHE_STRING_INDEX (STRING) VALUES %s ;",
		"(?)",
		makeArrayUnique(insertData.Strbuf),
	)
	if err == nil {
		_ = BulkInsertLimiter(
			"INSERT IGNORE INTO SEARCH_CACHE_ARTIST (`KEY`,BEATMAPSET_ID) VALUES %s ;",
			"((SELECT ID FROM  SEARCH_CACHE_STRING_INDEX WHERE `STRING` = ?), ?)",
			toIndexKV(insertData.Artist),
		)

		_ = BulkInsertLimiter(
			"INSERT IGNORE INTO SEARCH_CACHE_TITLE (`KEY`,BEATMAPSET_ID) VALUES %s ;",
			"((SELECT ID FROM  SEARCH_CACHE_STRING_INDEX WHERE `STRING` = ?), ?)",
			toIndexKV(insertData.Title),
		)
		_ = BulkInsertLimiter(
			"INSERT IGNORE INTO SEARCH_CACHE_CREATOR (`KEY`,BEATMAPSET_ID) VALUES %s ;",
			"((SELECT ID FROM  SEARCH_CACHE_STRING_INDEX WHERE `STRING` = ?), ?)",
			toIndexKV(insertData.Creator),
		)
		_ = BulkInsertLimiter(
			"INSERT IGNORE INTO SEARCH_CACHE_TAG (`KEY`,BEATMAPSET_ID) VALUES %s ;",
			"((SELECT ID FROM  SEARCH_CACHE_STRING_INDEX WHERE `STRING` = ?), ?)",
			toIndexKV(insertData.Tags),
		)
	}

}

func toIndexKV(data []row) (AA []interface{}) {
	for _, A := range data {
		for _, K := range A.KEY {
			AA = append(AA, K, A.BeatmapsetId)
		}
	}
	return
}

func stringArrayToInterfaceArray(sa *[]string) (ia []interface{}) {

	ia = make([]interface{}, len(*sa))
	for i, v := range *sa {
		ia[i] = v
	}
	return
}

func splitString(input string) (ss []string) {
	for _, s := range strings.Split(strings.ToLower(regexpReplace.ReplaceAllString(input, " ")), " ") {
		if s == "" || s == " " {
			continue
		}
		ss = append(ss, s, porter2.Stemmer.Stem(s))
	}
	return
}

func repeatStringArray(s string, count int) (arr []string) {
	for i := 0; i < count; i++ {
		arr = append(arr, s)
	}
	return
}
func makeArrayUnique(array []string) []interface{} {

	keys := make(map[string]struct{})
	res := make([]interface{}, 0)
	for _, s := range array {
		keys[s] = struct{}{}
	}
	for i := range keys {
		res = append(res, i)
	}
	return res
}