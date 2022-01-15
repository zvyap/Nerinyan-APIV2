package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pterm/pterm"
	"github.com/thftgr/osuFastCashedBeatmapMirror/Logger"
	"github.com/thftgr/osuFastCashedBeatmapMirror/Route"
	"github.com/thftgr/osuFastCashedBeatmapMirror/config"
	"github.com/thftgr/osuFastCashedBeatmapMirror/db"
	"github.com/thftgr/osuFastCashedBeatmapMirror/src"
	"log"
	"net/http"
	"os"
	"runtime"
)

// TODO DB 테이블 없으면 자동으로 생성하게
// TODO 로그 디비에 넣을때 어떤 데이터 넣을지.
// TODO 서버끼리 서로 비트맵파일 동기화 시킬수 있게
// TODO
// TODO
// TODO
func init() {
	ch := make(chan struct{})
	config.LoadConfig()
	src.StartIndex()
	db.ConnectMaria()
	go src.LoadBancho(ch)
	_ = <-ch
	if os.Getenv("debug") != "true" {
		go src.RunGetBeatmapDataASBancho()
	}
}

func main() {
	e := echo.New()
	e.HideBanner = true
	go func() {
		for {
			<-Logger.Ch
			e.Logger.SetOutput(log.Writer())
			pterm.Info.Println("UPDATED ECHO LOGGER.")
		}
	}()

	e.Pre(middleware.RemoveTrailingSlash())

	e.Use(
		middleware.Logger(),
		middleware.CORSWithConfig(middleware.CORSConfig{AllowOrigins: []string{"*"}, AllowMethods: []string{echo.GET, echo.HEAD}}),
		//middleware.RateLimiterWithConfig(middleWareFunc.RateLimiterConfig),
		middleware.RequestID(),
		middleware.Recover(),
	)
	// docs ============================================================================================================
	e.GET("/", Route.Wiki)

	// 서버상태 체크용 ====================================================================================================
	e.GET("/health", Route.Health)
	e.GET("/robots.txt", Route.Robots)
	e.GET("/api/status", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"Cpu Thread Count":     runtime.NumCPU(),
			"Running Thread Count": runtime.NumGoroutine(),
		})
	})

	// 맵 파일 다운로드 ===================================================================================================
	e.GET("/d/:id", Route.DownloadBeatmapSet)
	//TODO 맵아이디, 맵셋아이디 지원
	//e.GET("/d/:id", Route.DownloadBeatmapSet, middleWareFunc.LoadBalancer)

	// 비트맵 리스트 검색용 ================================================================================================
	e.GET("/api/search", Route.Search)
	e.GET("/api/search/beatmap/:mi", Route.SearchByBeatmapId)
	e.GET("/api/search/beatmapset/:si", Route.SearchByBeatmapSetId)

	// 서버 데이터 강제 업데이트용. ==========================================================================================
	// TODO 맵 굳이 한개씩 강제업데이트할 이유가 없음. 맵셋으로 업데이트만 지원
	e.GET("/api/update/beatmapset/:id", func(c echo.Context) error {

		//src.ManualUpdateBeatmapSet()
		return nil
	})

	pterm.Info.Println("ECHO STARTED AT", config.Setting.Port)
	e.Logger.Fatal(e.Start(":" + config.Setting.Port))

}

//var (
//	upgrader = websocket.Upgrader{}
//)

//func hello(c echo.Context) error {
//	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
//	if err != nil {
//		return err
//	}
//	defer ws.Close()
//
//	for {
//		// Write
//		err := ws.WriteMessage(websocket.TextMessage, []byte("Hello, Client!"))
//		if err != nil {
//			c.Logger().Error(err)
//		}
//
//		// Read
//		_, msg, err := ws.ReadMessage()
//		if err != nil {
//			c.Logger().Error(err)
//		}
//		fmt.Printf("%s\n", msg)
//	}
//}
