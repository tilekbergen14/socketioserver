package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
)

func main() {
	router := gin.New()
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		// s.SetContext("")
		fmt.Println("New connection")
		return nil
	})

	server.OnEvent("/", "msg", func(s socketio.Conn, msg string) {
		// fmt.Println("New message:", msg)
		// s.Emit("reply", msg)
		server.BroadcastToNamespace("", "reply", msg)
	})

	server.OnEvent("/chat", "msg", func(s socketio.Conn, msg string) string {
		s.SetContext(msg)
		return "recv " + msg
	})

	server.OnEvent("/", "bye", func(s socketio.Conn) string {
		last := s.Context().(string)
		s.Emit("bye", last)
		s.Close()
		return last
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		// server.Remove(s.ID())
		fmt.Println("meet error:", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {

		// Add the Remove session id. Fixed the connection & mem leak
		// server.Remove(s.ID())
		fmt.Println("closed =>", reason)
	})

	go func() {
		if err := server.Serve(); err != nil {
			log.Fatalf("socketio listen error: %s\n", err)
		}
	}()
	defer server.Close()

	// http.Handle("/socket.io/", server)
	// http.Handle("/", http.FileServer(http.Dir("./assets")))

	// log.Println("Serving at localhost:8000...")
	// log.Fatal(http.ListenAndServe(":8000", nil))

	router.Use(GinMiddleware("http://localhost:5173"))
	router.GET("/socket.io/*any", gin.WrapH(server))
	router.POST("/socket.io/*any", gin.WrapH(server))
	router.Static("/assets", "./assets")
	router.StaticFS("/more_static", http.Dir("my_file_system"))
	router.StaticFile("/favicon.ico", "./resources/favicon.ico")

	if err := router.Run(":8000"); err != nil {
		log.Fatal("failed run app: ", err)
	}
}

func GinMiddleware(allowOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Content-Length, X-CSRF-Token, Token, session, Origin, Host, Connection, Accept-Encoding, Accept-Language, X-Requested-With")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Request.Header.Del("Origin")

		c.Next()
	}
}
