// Пример использования webctl
package main

import (
	"fmt"
	"net"
	"strconv"
	"trueserver/webctl"
	"os"
	"os/signal"
	"time"
	"log"
)

func main() {
	// Инициализация web-сервера
	webctl.NeedExit = false; // флаг для завершения работы
	var web webctl.WebCtl
	webctl.GlobalConfig.SetManagerSrv("127.0.0.1", 4400)
	web.SetHost(net.ParseIP(webctl.GlobalConfig.ManagerSrvAddr()))
	web.SetPort(webctl.GlobalConfig.ManagerSrvPort())

	fmt.Println("Web control configured: " + "http://" + webctl.GlobalConfig.ManagerSrvAddr() + ":" + strconv.Itoa(int(webctl.GlobalConfig.ManagerSrvPort())))

	/* Запускаем сервер обслуживания WebCtl */
	err := web.StartServe()
	if err != nil {
		log.Println("HTTP сервер: Ошибка. ", err)
		os.Exit(1)
	}

	/* Перехват CTRL+C для завершения приложения */
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			fmt.Printf("\nReceived %v, shutdown procedure initiated.\n\n", sig)
			webctl.Quit <- 1
			webctl.NeedExit = true
		}
	}()

	// Цикл с таймером для ожидания команды завершения
	ticker := time.NewTicker(time.Second * 1) // Запускаем обработчик каждую секунду

	// Зациклимся с таймером посекундно пока не получим команду завершения работы.
	for range ticker.C {
		if !webctl.NeedExit {
			continue
		}
		break
	}
	ticker.Stop()
}

