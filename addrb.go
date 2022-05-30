package main

import (
  "fmt"
  "flag"
  "os"
  "strings"

  "github.com/emersion/go-vcard"
  "github.com/mrusme/addrb/dav"
  "github.com/mrusme/addrb/store"
)


func main() () {
  var username   string
  var password   string
  var endpoint   string
  var addrbDb    string

  var refresh    bool
  var lookupAttr string

  flag.StringVar(
    &username,
    "carddav-username",
    os.Getenv("CARDDAV_USERNAME"),
    "CardDAV username (HTTP Basic Auth)",
  )
  flag.StringVar(
    &password,
    "carddav-password",
    os.Getenv("CARDDAV_PASSWORD"),
    "CardDAV password (HTTP Basic Auth)",
  )
  flag.StringVar(
    &endpoint,
    "carddav-endpoint",
    os.Getenv("CARDDAV_ENDPOINT"),
    "CardDAV endpoint (HTTP(S) URL)",
  )
  flag.StringVar(
    &addrbDb,
    "database",
    os.Getenv("ADDRB_DB"),
    "Local vcard database",
  )

  flag.BoolVar(
    &refresh,
    "r",
    false,
    "Refresh local vcard database",
  )
  flag.StringVar(
    &lookupAttr,
    "l",
    vcard.FieldFormattedName,
    "Lookup attribute",
  )

  flag.Parse()

  args := flag.Args()

  if len(args) == 0 {
    flag.PrintDefaults()
    os.Exit(1)
  }

  lookupVal := strings.Join(args, " ")

  db, err := store.Open(addrbDb)
  if err != nil {
    fmt.Printf("%s\n", err)
    os.Exit(1)
  }
  defer db.Close()

  if refresh == true {
    cd, err := dav.New(endpoint, username, password)
    fmt.Printf("%s\n", err)
    err = cd.RefreshAddressBooks()
    fmt.Printf("%s\n", err)

    paths := cd.GetAddressBookPaths()
    fmt.Printf("%s\n", paths[0])
    vcs := cd.GetVcardsInAddressBook(paths[0])
    fmt.Printf("%d\n", len(vcs))

    err = db.Upsert(vcs)
    if err != nil {
      fmt.Printf("%s\n", err)
    }
  }

  foundVcs, err := db.FindBy(lookupAttr, lookupVal)
  if err != nil {
    fmt.Printf("%s\n", err)
  }

  fmt.Printf("FOUND:\n")
  for _, vc := range foundVcs {
    fmt.Printf("%s: %s\n", vc.Get(vcard.FieldUID).Value, vc.PreferredValue(vcard.FieldFormattedName))
  }

  os.Exit(0)
}

