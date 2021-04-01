// NetD makes network device operations easy.
// Copyright (C) 2019  sky-cloud.net
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sky-cloud-tec/netd/api/routers"
	"github.com/sky-cloud-tec/netd/common"
	"github.com/sky-cloud-tec/netd/ingress"

	"github.com/songtianyi/rrframework/logs"
	"github.com/urfave/cli"
)

// AppConfig app configurations
type AppConfig struct {
	logCfg *common.LogConfig
}

var appConfig *AppConfig

func init() {
	appConfig = &AppConfig{
		logCfg: &common.LogConfig{},
	}
	if common.AppConfigInstance == nil {
		common.AppConfigInstance = &common.AppConfig{
			Confidence: 30,
		}
	}
}

func initLogger() error {
	// set logger
	property := `{"filename": "` + appConfig.logCfg.Filepath +
		`", "maxlines" : 10000000, "maxsize": ` + strconv.Itoa(appConfig.logCfg.MaxSize) + `}`
	fmt.Println(property, appConfig.logCfg)
	logs.SetLevel(common.MapStringToLevel[appConfig.logCfg.Level])
	return logs.SetLogger("file", property)

}

func jrpcHandler(c *cli.Context) error {
	// init logger
	if err := initLogger(); err != nil {
		return err
	}
	go func() {
		if err := routers.SetupRouter(c.String("api-addr")).Run(c.String("api-addr")); err != nil {
			panic(err)
		}
	}()
	common.AppConfigInstance.LogCfgDir = strings.TrimSuffix(common.AppConfigInstance.LogCfgDir, "/")
	// init jrpc
	jrpc, _ := ingress.NewJrpc(c.String("addr"))
	jrpc.Register(new(ingress.CliHandler))
	if err := jrpc.Serve(); err != nil {
		return err
	}
	return nil
}

func main() {
	app := cli.NewApp()
	app.Usage = `NetD make network device operations easy!
	It's a dammon app which allow you to run cli commands through grpc, amqp(not support yet) etc.`
	app.Version = "2.0.0"
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		{
			Name:  "songtianyi",
			Email: "songtianyi@sky-cloud.net",
		},
	}
	app.Copyright = "Copyright (c) 2017-2019 sky-cloud.net"
	app.Commands = []cli.Command{
		{
			Name:    "jrpc",
			Aliases: []string{"jrpc"},
			Usage:   "Run netd with jrpc ingress",
			Action:  jrpcHandler,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "address, addr",
					Value: "0.0.0.0:8188", // default port 8188
					Usage: "jprc listen address",
				},
				cli.StringFlag{
					Name:  "api-address, api-addr",
					Value: "0.0.0.0:8189",
					Usage: "api listen address",
				},
				cli.IntFlag{
					Name:        "confidence, ce",
					Value:       30,
					Usage:       "encoding convert confidence",
					Required:    false,
					Destination: &common.AppConfigInstance.Confidence,
				},
				cli.IntFlag{
					Name:        "log-cfg-flag, lcf",
					Value:       0, // false
					Usage:       "write command output cfg to file as binary",
					Required:    false,
					Destination: &common.AppConfigInstance.LogCfgFlag,
				},
				cli.StringFlag{
					Name:        "log-cfg-dir, lcd",
					Value:       "/var/log/netd",
					Required:    false,
					Destination: &common.AppConfigInstance.LogCfgDir,
				},
			},
		},
		{
			Name:    "hotfix",
			Aliases: []string{"hotfix"},
			Usage:   "Run hotfix cli to fix regex patterns online\n\t\t\tplease note, fixed pattern will lost when netd restart.",
		},
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "logfile, lf",
			Value:       "/var/log/netd/netd.log",
			Usage:       "logfile path",
			Destination: &appConfig.logCfg.Filepath,
		},
		cli.StringFlag{
			Name:        "loglevel, ll",
			Value:       "INFO",
			Usage:       "log level, EMERGENCY|ALERT|CRITICAL|ERROR|WARNING|NOTICE|INFO|DEBUG",
			Destination: &appConfig.logCfg.Level,
		},
		cli.IntFlag{
			Name:        "maxsize, ms",
			Value:       10240000, // default log file max size 10M
			Usage:       "log file max size",
			Destination: &appConfig.logCfg.MaxSize,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
