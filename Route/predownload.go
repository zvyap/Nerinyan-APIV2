package Route

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/Nerinyan/Nerinyan-APIV2/Logger"
	"github.com/Nerinyan/Nerinyan-APIV2/banchoCroller"
	"github.com/Nerinyan/Nerinyan-APIV2/bodyStruct"
	"github.com/Nerinyan/Nerinyan-APIV2/config"
	"github.com/Nerinyan/Nerinyan-APIV2/db"
	"github.com/Nerinyan/Nerinyan-APIV2/src"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
	"io"
	"net/http"
	"strconv"
	"time"
)

func Predownload(c echo.Context) (err error) {

	//1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False.
	noVideo, _ := strconv.ParseBool(c.QueryParam("noVideo"))
	noVideo2, _ := strconv.ParseBool(c.QueryParam("nv"))
	noVideo = noVideo || noVideo2

	mid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, logger.Error(c, &bodyStruct.ErrorStruct{
			Code:      "DownloadBeatmapSet-001",
			Path:      c.Path(),
			RequestId: c.Response().Header().Get("X-Request-ID"),
			Error:     err,
			Message:   "request parse error",
		}))
	}

	go banchoCroller.ManualUpdateBeatmapSet(mid)

	row := db.Maria.QueryRow(`SELECT beatmapset_id,artist,title,last_updated,video FROM osu.beatmapset WHERE beatmapset_id = ?`, mid)
	err = row.Err()
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, logger.Error(c, &bodyStruct.ErrorStruct{
				Code:      "DownloadBeatmapSet-002",
				Path:      c.Path(),
				RequestId: c.Response().Header().Get("X-Request-ID"),
				Error:     err,
				Message:   "not in database",
			}))
		}
		return c.JSON(http.StatusInternalServerError, logger.Error(c, &bodyStruct.ErrorStruct{
			Code:      "DownloadBeatmapSet-003",
			Path:      c.Path(),
			RequestId: c.Response().Header().Get("X-Request-ID"),
			Error:     err,
			Message:   "database Query error",
		}))
	}

	var a struct {
		Id          string
		Artist      string
		Title       string
		LastUpdated string
		Video       bool
	}

	if err = row.Scan(&a.Id, &a.Artist, &a.Title, &a.LastUpdated, &a.Video); err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, logger.Error(c, &bodyStruct.ErrorStruct{
				Code:      "DownloadBeatmapSet-004",
				Path:      c.Path(),
				RequestId: c.Response().Header().Get("X-Request-ID"),
				Error:     err,
				Message:   "not in database",
			}))

		}
		return c.JSON(http.StatusInternalServerError, logger.Error(c, &bodyStruct.ErrorStruct{
			Code:      "DownloadBeatmapSet-005",
			Path:      c.Path(),
			RequestId: c.Response().Header().Get("X-Request-ID"),
			Error:     err,
			Message:   "database Query error",
		}))
	}

	lu, err := time.Parse("2006-01-02 15:04:05", a.LastUpdated)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, logger.Error(c, &bodyStruct.ErrorStruct{
			Code:      "DownloadBeatmapSet-006",
			Path:      c.Path(),
			RequestId: c.Response().Header().Get("X-Request-ID"),
			Error:     err,
			Message:   "time Parse error",
		}))
	}

	url := fmt.Sprintf("https://osu.ppy.sh/api/v2/beatmapsets/%d/download", mid)
	if a.Video && noVideo {
		mid *= -1
		a.Title += " [no video]"
		url += "?noVideo=1"
	}

	serverFileName := fmt.Sprintf("%s/%d.osz", config.Config.TargetDir, mid)

	if src.FileList[mid].Unix() >= lu.Unix() { // 맵이 최신인경우
		return c.Response().NoContent(http.StatusOK)
	}

	//==========================================
	//=        비트맵 파일이 서버에 없는경우        =
	//==========================================

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, logger.Error(c, &bodyStruct.ErrorStruct{
			Code:      "DownloadBeatmapSet-007",
			Path:      c.Path(),
			RequestId: c.Response().Header().Get("X-Request-ID"),
			Error:     err,
			Message:   "Bancho request Build Error",
		}))
	}
	req.Header.Add("Authorization", config.Config.Osu.Token.TokenType+" "+config.Config.Osu.Token.AccessToken)

	res, err := client.Do(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, logger.Error(c, &bodyStruct.ErrorStruct{
			Code:      "DownloadBeatmapSet-008",
			Path:      c.Path(),
			RequestId: c.Response().Header().Get("X-Request-ID"),
			Error:     err,
			Message:   "Bancho request Build Erro",
		}))
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return c.JSON(http.StatusInternalServerError, logger.Error(c, &bodyStruct.ErrorStruct{
			Code:      "DownloadBeatmapSet-009",
			Path:      c.Path(),
			RequestId: c.Response().Header().Get("X-Request-ID"),
			Error:     errors.New(http.StatusText(res.StatusCode)),
			Message:   "Bancho request Error. :" + res.Status,
		}))
	}

	pterm.Info.Println("beatmapSet Downloading at", serverFileName)

	cLen, _ := strconv.Atoi(res.Header.Get("Content-Length"))
	c.Response().Header().Set("Content-Length", res.Header.Get("Content-Length"))
	c.Response().Header().Set("Content-Disposition", res.Header.Get("Content-Disposition"))
	c.Response().Header().Set("Content-Type", "application/x-osu-beatmap-archive")

	var buf bytes.Buffer

	//http://localhost/d/1573058
	//http://localhost/d/1469677

	defer c.Response().Flush()
	for i := 0; i < cLen; { // 읽을 데이터 사이즈 체크
		var b = make([]byte, 64000) // 바이트 배열
		n, e := res.Body.Read(b)    // 반쵸 스트림에서 64k 읽어서 바이트 배열 b 에 넣음

		i += n // 현재까지 읽은 바이트
		if n > 0 {
			buf.Write(b[:n]) // 서버에 저장할 파일 버퍼에 쓴다
		}

		if e == io.EOF {
			break
		} else if e != nil { //에러처리
			return c.JSON(http.StatusInternalServerError, logger.Error(c, &bodyStruct.ErrorStruct{
				Code:      "DownloadBeatmapSet-011",
				Path:      c.Path(),
				RequestId: c.Response().Header().Get("X-Request-ID"),
				Error:     e,
				Message:   "fail to read Bancho Stream",
			}))
		}
	}
	if cLen == buf.Len() {
		return saveLocal(&buf, serverFileName, mid)
	}
	errMsg := fmt.Sprintf("filesize not match: bancho response bytes : %d | downloaded bytes : %d", cLen, buf.Len())
	pterm.Error.Printfln(errMsg)
	return errors.New(errMsg)

}
//
