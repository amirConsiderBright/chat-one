package main

import (
	model "char/core"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
)

func init() {
	logFile, err := os.OpenFile("./ws.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile) // 将文件设置为log输出的文件
	log.SetPrefix("TRACE: ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Llongfile)
}

//const (
//	QueryUsers = iota
//)

//获取全部用户
func getUsers() []Friend {
	var friends = make([]Friend, 0, len(connMap))
	for s, user := range connMap {
		friends = append(friends, Friend{
			Token: s,
			Name:  user.ChineseName,
		})
	}
	return friends
}

type Friend struct {
	Token string
	Name  string
}

var connMap = make(map[string]model.User)
var closeChan = make(chan string, 100)

func main() {
	closeChan := make(chan string, 100)
	go removeUser(closeChan)
	http.HandleFunc("/ws", Upgrade)
	http.HandleFunc("/getFriend", GetFriend)
	log.Println("准备启动")
	err := http.ListenAndServe(":5900", nil)
	if err != nil {
		log.Println(err)
		return
	}
}
func removeUser(closeChan chan string) {
	for {
		token := <-closeChan
		delete(connMap, token)
		if user, ok := connMap[token]; ok {
			delete(connMap, token)
			log.Printf("删除成功token:%s,name:%s\n", token, user.ChineseName)
		}
	}
}
func GetFriend(writer http.ResponseWriter, _ *http.Request) {
	_, err := writer.Write([]byte("成功"))
	if err != nil {
		return

	}
}

func Upgrade(w http.ResponseWriter, r *http.Request) {
	//r.Cookie("name")
	//log.Println(r.URL.String())

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	user := model.NewUser(conn, "qiang")
	log.Println("接入用户:", user)
	connMap[user.Token] = user
	go func(token string, closeChan chan<- string) {
		for {
			//var msg model.Mess
			messageType, str, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				closeChan <- token
				return
			}
			switch messageType {
			case websocket.CloseMessage:
				closeChan <- token
			case websocket.TextMessage:
				log.Println(string(str))
			default:
				log.Println("接收到无意义数据:", string(str))
			}
		}
	}(user.Token, closeChan)
}
