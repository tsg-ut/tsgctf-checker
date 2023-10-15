package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tsg-ut/tsgctf-checker/badge"
	"github.com/tsg-ut/tsgctf-checker/checker"
	"go.uber.org/zap"
)

func get_port() int {
	// priority is command-line > ENVVAR.
	var port_num int = 0
	port := flag.Int("port", 8080, "Port number this badge server listens to. (can be specified also by $BADGEPORT envvar.)")
	flag.Parse()

	port_num = *port
	port_str := os.Getenv("BADGE_PORT")
	if len(port_str) != 0 {
		if port_num_tmp, err := strconv.Atoi(port_str); err == nil {
			port_num = port_num_tmp
		}
	}

	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "port":
			port_num = *port
			break
		}
	})

	return port_num
}

func main() {
	// force running in production mode
	gin.SetMode(gin.ReleaseMode)

	// init logger
	level := zap.NewAtomicLevel()
	level.SetLevel(zap.DebugLevel)
	slogger, _ := zap.NewDevelopment()
	defer slogger.Sync()
	logger := slogger.Sugar()

	// get Badger
	db, err := checker.Connect(os.Getenv("DBUSER"), os.Getenv("DBPASS"), os.Getenv("DBHOST"), os.Getenv("DBNAME"))
	if err != nil {
		logger.Fatal(err)
	}
	badger := badge.NewBadger(db)

	// init server
	server := gin.Default()

	// health check EP
	server.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// badge EP
	server.GET("/badge/:chall_name", func(c *gin.Context) {
		chall_name := c.Params.ByName("chall_name")

		badge_url, err := badger.GetBadge(chall_name)
		if err != nil {
			logger.Warnf("%v", err)
			c.String(http.StatusInternalServerError, "Something went to bad when fetching test result.")
			return
		}

		c.Header("Cache-Control", "max-age=60, public, immutable, must-revalidate")
		c.Redirect(http.StatusFound, badge_url)
		return
	})

	// default error badge
	server.GET("/badge/error", func(c *gin.Context) {
		c.Header("Cache-Control", "max-age=60, public, immutable, must-revalidate")
		c.Redirect(http.StatusFound, "https://img.shields.io/badge/error-status_fetching_fails-red")
	})

	// run server
	port := get_port()
	port_str := fmt.Sprintf(":%v", port)
	logger.Infof("Badge server running on %s.", port_str)
	server.Run(port_str)
}
