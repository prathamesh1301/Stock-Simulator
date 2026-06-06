package domain

import gorilla "github.com/gorilla/websocket"


type Client struct {
	Conn *gorilla.Conn
	Send chan []byte
	Subscriptions map[string]bool
}