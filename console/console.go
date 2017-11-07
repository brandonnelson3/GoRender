package console

import (
	"html/template"
	"log"
	"net/http"

	"github.com/brandonnelson3/GoRender/messagebus"
	"github.com/labstack/echo"
	"golang.org/x/net/websocket"
)

// WebsocketMessage is representative of a single message being sent from the backend to the frontend, or the other direction.
type WebsocketMessage struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// InitConsole  starts the dashboard's webserver in a go routine.
func InitConsole() {
	go func() {
		t := &Template{
			templates: template.Must(template.ParseGlob("console/views/*.html")),
		}
		e := echo.New()
		e.Static("/static", "console/static")
		e.Renderer = t
		e.GET("/ws", websocketHandler)
		e.GET("/", func(c echo.Context) error {
			return c.Render(http.StatusOK, "dashboard", nil)
		})
		e.Logger.Fatal(e.Start(":8080"))
	}()
}

func websocketHandler(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		log.Printf("Websocket connected...")
		defer ws.Close()

		messagebus.RegisterType("console",
			func(m *messagebus.Message) {
				websocketMessage := &WebsocketMessage{
					Type:  m.Data1.(string),
					Value: m.Data2.(string),
				}
				if err := websocket.JSON.Send(ws, websocketMessage); err != nil {
					log.Printf("Send error from websocket: %v", err)
				}
			})
		for {
			websocketMessage := WebsocketMessage{}
			if err := websocket.JSON.Receive(ws, websocketMessage); err != nil {
				log.Printf("Recieve error from websocket: %v", err)
				return
			}

			log.Printf("Got message: %v", websocketMessage)
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}
