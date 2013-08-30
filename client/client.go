package client

import (
    "fmt"
    "os"
    "github.com/jaekwon/go-prelude/colors"
)

var HeaderLine string = "Gourami "+colors.Cyan("<°", colors.Red("\\"), "\\", colors.Red("\\"), "<")+" (version 0.0)\n"

func PrintHelp() {
    fmt.Println("commands: generate\n")
}

func Main() {
    fmt.Println(HeaderLine)
    if len(os.Args) == 1 {
        PrintHelp()
        return
    }

    args := os.Args[1:]
    if args[0] == "generate" {
        config, err := GenerateConfig()
        if err != nil {
            fmt.Println(colors.Red("Error: " + err.Error()))
            return }
        err = config.Save("./config", "password")
        if err != nil {
            fmt.Println(colors.Red("Error: " + err.Error()))
            return }
    } else {
        PrintHelp()
    }
}
