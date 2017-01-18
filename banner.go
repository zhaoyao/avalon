package main

import (
	"fmt"
)

var bannerText = `
   _____               .__
  /  _  \___  _______  |  |   ____   ____
 /  /_\  \  \/ /\__  \ |  |  /  _ \ /    \
/    |    \   /  / __ \|  |_(  <_> )   |  \
\____|__  /\_/  (____  /____/\____/|___|  /
        \/           \/                 \/
`

func printBanner() {
	fmt.Println(bannerText)
}
