package main

import (
	"fmt"
	"os"
	"time"

	"nfa-tool/internal/steam"
	"nfa-tool/internal/token"
)

func main() {
	tokens, err := steam.HarvestConnectCache()
	if err != nil {
		fmt.Println("harvest error:", err)
		os.Exit(1)
	}
	if len(tokens) == 0 {
		fmt.Println("no tokens in ConnectCache")
		os.Exit(1)
	}
	for name, tok := range tokens {
		info, err := token.ParseAndValidate(tok)
		if err != nil {
			fmt.Printf("%s: parse error: %v\n", name, err)
			continue
		}
		fmt.Printf("%s: steamID=%s expires=%s\n", name, info.SteamID, info.ExpiresAt)
		eres, err := steam.CheckTokenCM(tok, 20*time.Second)
		if err != nil {
			fmt.Printf("%s: cm error: %v\n", name, err)
			continue
		}
		fmt.Printf("%s: eresult=%d\n", name, eres)
	}
}
