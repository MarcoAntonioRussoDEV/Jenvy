package ui

import (
	"fmt"

	"github.com/mbndr/figlet4go"
)

func ShowBanner() {
    render := figlet4go.NewAsciiRender()
    options := figlet4go.NewRenderOptions()
    options.FontName = "standard" // puoi cambiare con "block", "slant", ecc.

    output, err := render.RenderOpts("Java Version Manager", options)
    if err != nil || output == "" {
        fmt.Println("🚀 Java Version Manager – intelligent JDK explorer")
    } else {
        fmt.Println(output)
    }
    fmt.Println("🔍 Matching by tag  |  💡 LTS-first logic")
}
