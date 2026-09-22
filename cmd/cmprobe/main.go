package main

import (
	"fmt"
	"os"
	"time"

	"nfa-tool/internal/steam"
)

var eresultNames = map[int]string{
	1:  "OK",
	2:  "Fail",
	3:  "NoConnection",
	5:  "InvalidPassword",
	6:  "LoggedInElsewhere",
	8:  "InvalidParam",
	15: "AccessDenied",
	20: "ServiceUnavailable",
	27: "Expired",
	48: "TryAnotherCM",
	63: "AccountLogonDenied",
	65: "InvalidLoginAuthCode",
	73: "AccountLockedDown",
	84: "RateLimitExceeded",
	88: "TwoFactorCodeMismatch",
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: cmprobe <refresh-token>")
		os.Exit(2)
	}
	token := os.Args[1]

	eresult, err := steam.CheckTokenCM(token, 15*time.Second)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	name := eresultNames[eresult]
	if name == "" {
		name = "unknown"
	}
	fmt.Printf("eresult=%d (%s)\n", eresult, name)
}
