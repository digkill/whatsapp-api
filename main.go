package main

import (
	"context"
	_ "database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3" // Импорт с побочным эффектом
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
	"net/http"
)

type MessageRequest struct {
	Number  string `json:"number"`
	Message string `json:"message"`
}

func eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		fmt.Println("Received a message!", v.Message.GetConversation())
	}
}

func main() {
	r := gin.Default()
	r.POST("/sendMessage", sendMessage)
	err := r.Run(":8081")
	if err != nil {
		panic(err)
	}
}

func whatsapp(phone string, message string) {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, err := sqlstore.New("sqlite3", "file:store.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}

	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}
	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	//client.AddEventHandler(eventHandler)

	if client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				// e.g. qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal
				fmt.Println("QR code:", evt.Code)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		err = client.Connect()
		if err != nil {
			panic(err)
		}

		//jid := jid.NewJID()

		jid := types.NewJID(phone, types.DefaultUserServer)

		// jid := types.NewJID("79194738112", "s.whatsapp.net")

		/*_, err = client.SendMessage(context.Background(), jid, &waE2E.Message{
			Conversation: proto.String("Hello, World! Blyaaa!!!"),
		})*/

		_, err = client.SendMessage(context.Background(), jid, &waE2E.Message{
			Conversation: proto.String(message),
		})

		if err != nil {
			panic(err)
		}

	}

	/*c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c*/

	client.Disconnect()
}

func sendMessage(c *gin.Context) {
	var req MessageRequest

	fmt.Println(req)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	whatsapp(req.Number, req.Message)

	c.JSON(http.StatusOK, gin.H{"status": "message sent"})
}
