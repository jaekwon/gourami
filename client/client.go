package client

import (
    "fmt"
    "os"
    "io/ioutil"
    "github.com/jaekwon/go-prelude/colors"
    "github.com/jaekwon/gourami/types"
)

var HeaderLine string = "Gourami "+colors.Cyan("<°", colors.Red("\\"), "\\", colors.Red("\\"), "<")+" (version 0.0)\n"

func GenerateIdentity() {
    filename := "./config"
    if _, err := os.Stat(filename); !os.IsNotExist(err) {
        configB, err := ioutil.ReadFile(filename)
        if err != nil {
            fmt.Println(colors.Red(err)+"\n")
            return
        }
        config := string(configB)
        if config != "" {
            fmt.Println(colors.Red("config file already exists\n"))
            return
        }
    }

    fmt.Println("generating identity...")

    identity := types.GenerateIdentity()

    ioutil.WriteFile("./config", []byte("something"), 0600)
}

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
        GenerateIdentity()
    } else {
        PrintHelp()
    }
}
