package main

import (
    "flag"
    "fmt"
    "os"
)

func usage() {
    fmt.Println("pgsdbuild - PGSD build tool (prototype)")
    fmt.Println()
    fmt.Println("Usage:")
    fmt.Println("  pgsdbuild image <image-id>")
    fmt.Println("  pgsdbuild iso <variant-id>")
    fmt.Println("  pgsdbuild list-images")
    fmt.Println()
}

func main() {
    flag.Usage = usage
    flag.Parse()

    args := flag.Args()
    if len(args) < 1 {
        usage()
        os.Exit(1)
    }

    cmd := args[0]
    switch cmd {
    case "image":
        if len(args) < 2 {
            fmt.Println("missing image-id")
            os.Exit(1)
        }
        id := args[1]
        fmt.Printf("[prototype] would build image %q here\n", id)
    case "iso":
        if len(args) < 2 {
            fmt.Println("missing variant-id")
            os.Exit(1)
        }
        v := args[1]
        fmt.Printf("[prototype] would build bootenv ISO %q here\n", v)
    case "list-images":
        fmt.Println("[prototype] no images yet; define images/*.lua to enable listing")
    default:
        fmt.Println("unknown command:", cmd)
        usage()
        os.Exit(1)
    }
}
