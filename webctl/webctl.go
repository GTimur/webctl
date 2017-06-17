// Реализует http-сервер с возможностью корректного завершения.
package webctl

import (
	"log"
	"fmt"
	"time"
	"net"
	"net/http"
	"context"
	"html/template"
)

type WebCtl struct {
	host     net.IP
	port     uint16
	islisten bool
}

type Config struct {
	managerSrv managerSrv
}

// Представляет адрес сервера управления программой и порт
type managerSrv struct {
	Addr string
	Port uint16
}

// Типизируем страницу для передачи данных в шаблон
type Page struct {
	Title   string
	Body    template.HTML
	LnkHome string
	DateNow template.HTML
}

var (
	GlobalConfig Config = Config{}        // Глобальная переменная для хранения настроек
	NeedExit bool                // Флаг для завершения работы сервера
	Quit = make(chan int, 1)        // Канал для завершения сервера HTTP
)

/*Сервер*/
//Запускает goroutine http.Server
func (w *WebCtl) StartServe() (err error) {
	//signal.Notify(Quit, os.Interrupt)
	srv := &http.Server{Addr : w.connString(),
		Handler: http.DefaultServeMux,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// для отдачи сервером статичных файлов из папки /static
	fs := http.FileServer(http.Dir("./static/"))
	//http.Handle("/static/", http.StripPrefix("/static/", fs))

	cssFileServer := http.StripPrefix("/static/", fs)
	http.Handle("/static/", cssFileServer)
	http.HandleFunc("/", urlhome) //Домашняя страница

	go func() {
		log.Println("Starting HTTP-server...")
		log.Fatalln("WebCtl error:", srv.ListenAndServe())
	}()

	go func() {
		<-Quit
		fmt.Println("Shutting down HTTP-server...")
		ctx, _ := context.WithTimeout(context.Background(), 1 * time.Second)
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalln("HTTP Shutdown error:", err)
		}
	}()
	w.islisten = true
	return err
}

//Функции установки значений
func (w *WebCtl) SetHost(host net.IP) {
	w.host = host
}

func (w *WebCtl) SetPort(port uint16) {
	w.port = port
}

/**/
func (w WebCtl) connString() string {
	return fmt.Sprintf("%s:%d", w.host.String(), w.port)
}

func (c *Config) SetManagerSrv(addr string, port uint16) {
	c.managerSrv = managerSrv{
		Addr: addr,
		Port: port,
	}
}

func (c *Config) ManagerSrvAddr() string {
	return c.managerSrv.Addr
}

func (c *Config) ManagerSrvPort() uint16 {
	return c.managerSrv.Port
}

/****/
//Обработчик запросов для home - пример
func urlhome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	body := `<h1>Welcome to homepage</h1>
	<p>You are welcome!</p>`

	main := `<!DOCTYPE html>
	<html>
	<head>
    	<meta charset="utf-8">
    	<meta http-equiv="X-UA-Compatible" content="IE=edge">
    	<meta name="viewport" content="width=device-width, initial-scale=1">
    	<title>{{ .Title }}</title>
	</head>
	<body>
		{{ .Body }}
	</body>
	</html>`

	page := Page{Title:"HOME PAGE",
		Body: template.HTML(body),
		LnkHome: "none",
		DateNow: "",
	}

	home_template := template.Must(template.New("main").Parse(main))

	if r.Method == "GET" {
		if err := home_template.ExecuteTemplate(w, "main", page); err != nil {
			fmt.Sprintln("Homepage handling error:", err.Error())
		}
		fmt.Println("Homepage: GET request.")
	} else {
		fmt.Println("Homepage: POST request.")
	}

}

/****/