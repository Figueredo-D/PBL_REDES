package main

import (
	"log"

	"github.com/Figueredo-D/PBL_REDES/internal/server"
)

func main() {
	srv := server.NewServer(":8080")
	if err := srv.Start(); err != nil {
		log.Fatalf("[FATAL] Erro ao rodar servidor: %v", err)
	}
}