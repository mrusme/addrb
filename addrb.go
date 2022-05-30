package main

import (
  "fmt"
  "os"

  "github.com/emersion/go-vcard"
  "github.com/mrusme/addrb/dav"
  "github.com/mrusme/addrb/store"
)


func main() {
  username := os.Getenv("CARDDAV_USERNAME")
  password := os.Getenv("CARDDAV_PASSWORD")
  endpoint := os.Getenv("CARDDAV_ENDPOINT")

  db, err := store.Open(os.Getenv("ADDRB_DB"))
  if err != nil {
    fmt.Printf("%s\n", err)
    return
  }
  defer db.Close()

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

  foundVcs, err := db.FindBy(vcard.FieldFormattedName, "Johnny Bravo")
  if err != nil {
    fmt.Printf("%s\n", err)
  }

  fmt.Printf("FOUND:\n")
  for _, vc := range foundVcs {
    fmt.Printf("%s: %s\n", vc.Get(vcard.FieldUID).Value, vc.PreferredValue(vcard.FieldFormattedName))
  }

  return
}

