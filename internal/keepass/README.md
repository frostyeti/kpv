# keepass

## Overview

The keepss mdoule wraps the gokeepasslib to make gokeepasslib easier to use in Go projects.  

Gokeepasslib works well, however, it requires a bit of boilerplate to use.

## Usage

To use `keepass`, import the module in your Go project:

```go
import "github.com/hyprxlabs/go/keepass"   

func main() {
    wd, _ := os.Getwd()
    dbPath := filepath.Join(wd, "save_test.kdbx")

    pwd := stringPtr("testpassword")

    options := keepass.KdbxOptions{
        Path:      dbPath,
        Secret:    pwd,
        Create:    true, // Create a new Kdbx file if it doesn't exist
        CreateDir: true, // Ensure directory is created if it doesn't exist
    }

    kdbx := keepass.New(options)
    require.NotNil(t, kdbx)
    err := kdbx.Create()
    if err != nil {
        log.Fatalf("Failed to create Kdbx file: %v", err)
        os.Exit(1)
    }

    _, err = keepass.Open(options)
    if err != nil {
        log.Fatalf("Failed to open Kdbx file: %v", err)
        os.Exit(1)
    }

    // Root Group
    rg := kdbx.Root()

    first := keepass.NewEntry()
    first.SetTitle("First Entry")
    first.SetPassword("firstpass")
    rg.AddEntry(first)

    second := keepass.NewEntry()
    second.SetTitle("Second Entry")
    second.SetPassword("secondpass")
    rg.AddEntry(second)

    err = kdbx.Save()

    assert.NoError(t, err)

    kdbx2, err := keepass.Open(keepass.KdbxOptions{
        Path:   dbPath,
        Secret: pwd,
    })

    if err != nil {
        log.Fatalf("Failed to open Kdbx file: %v", err)
        os.Exit(1)
    }

    if kdbx2 == nil {
        log.Fatal("Kdbx should not be nil after successful open")
        os.Exit(1)
    }

    if !kdbx2.IsOpen() {
        log.Fatal("Kdbx should be open after successful save")
        os.Exit(1)
    }

    // uses a path delimiter to find the entry by group name and entry title
    // e.g. "group1/nestedGroup/First Entry"
    // you do not need to specify the root group
    firstResult := kdbx2.FindEntry("First Entry")

    if firstResult == nil {
        log.Fatal("First entry should not be nil after save")
        os.Exit(1)
    }

    println("First Entry Title:", firstResult.GetTitle())
    println("First Entry Password:", firstResult.GetPassword())
}
```
