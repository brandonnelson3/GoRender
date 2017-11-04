package main

import (
	"fmt"

	"github.com/brandonnelson3/GoRender/messagebus"
)

func logger(m *messagebus.Message) {
	fmt.Println(m.System + ": " + m.Data1.(string))
}

func init() {
	messagebus.RegisterType("log", logger)
}
