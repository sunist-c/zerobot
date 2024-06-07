// Package main ZeroBot-Plugin main file
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"github.com/FloatTech/ZeroBot-Plugin/cls"
	_ "github.com/FloatTech/ZeroBot-Plugin/console" // 更改控制台属性
	"github.com/FloatTech/ZeroBot-Plugin/manager"
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ceobebot"
	"github.com/FloatTech/floatbox/process"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"os"
	// webctrl "github.com/FloatTech/zbputils/control/web"
	// -----------------------以上为内置依赖，勿动------------------------ //
)

type zbpcfg struct {
	Z zero.Config        `json:"zero"`
	W []*driver.WSClient `json:"ws"`
	S []*driver.WSServer `json:"wss"`
}

var config zbpcfg

func init() {
	// 解析命令行参数
	zbpConfig := flag.String("c", "", "Run from config file.")
	managerCfg := flag.String("m", "", "Manager config file.")
	globalCfg := flag.String("g", "", "Plugin public config file.")
	flag.Parse()

	if *zbpConfig == "" || *managerCfg == "" || *globalCfg == "" {
		panic(errors.New("zerobot 的 ceobebot 分支不能在没有配置文件的情况下运行，必须同时配置 -c -m -g 参数"))
	}

	f, err := os.Open(*zbpConfig)
	if err != nil {
		panic(err)
	}

	config.W = make([]*driver.WSClient, 0, 2)
	err = json.NewDecoder(f).Decode(&config)
	f.Close()
	if err != nil {
		panic(err)
	}
	config.Z.Driver = make([]zero.Driver, len(config.W)+len(config.S))
	for i, w := range config.W {
		config.Z.Driver[i] = w
	}
	for i, s := range config.S {
		config.Z.Driver[i+len(config.W)] = s
	}

	err = manager.LoadPublicConfig(*globalCfg)
	if err != nil {
		panic(err)
	}

	manager.SetLogger(cls.Logger())
	manager.Initialize(*managerCfg)
}

func main() {
	zero.RunAndBlock(&config.Z, process.GlobalInitMutex.Unlock)
}
