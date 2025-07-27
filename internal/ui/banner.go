package ui

import (
	"fmt"

	"jenvy/internal/utils"

	"github.com/mbndr/figlet4go"
)

func ShowBanner() {
	render := figlet4go.NewAsciiRender()
	options := figlet4go.NewRenderOptions()
	options.FontName = "standard" // puoi cambiare con "block", "slant", ecc.

	output, err := render.RenderOpts("Jenvy", options)
	if err != nil || output == "" {
		fmt.Println(utils.ColorText("[JENVY] Developer Kit Manager - intelligent OpenJDK explorer", utils.BrightCyan))
	} else {
		fmt.Print(utils.ColorText(output, utils.BrightBlue))
	}
	fmt.Println(utils.SearchText("Matching by tag") + "  |  " + utils.ColorText("[LTS]", utils.BrightGreen) + " LTS-first logic")
}
