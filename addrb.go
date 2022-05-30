package main

import (
  "fmt"
  "os"

  "github.com/emersion/go-vcard"
  "github.com/emersion/go-webdav"
  "github.com/emersion/go-webdav/carddav"

)


func main() {
  username := os.Getenv("CARDDAV_USERNAME")
  password := os.Getenv("CARDDAV_PASSWORD")
  endpoint := os.Getenv("CARDDAV_ENDPOINT")

  httpClient := webdav.HTTPClientWithBasicAuth(nil, username, password)

  cd, err := carddav.NewClient(httpClient, endpoint)
  if err != nil {
    fmt.Printf("%s\n", err)
    return
  }

  s, err := cd.FindAddressBookHomeSet(fmt.Sprintf("principals/%s", username))
  fmt.Printf("%s\n%s\n", s, err)

  ab, err := cd.FindAddressBooks(s)
  fmt.Printf("%v\n%s\n", ab[0].Path, err)

  multiGet := new(carddav.AddressBookQuery)
  multiGet.DataRequest = carddav.AddressDataRequest{
    Props: []string{
      vcard.FieldEmail,
      vcard.FieldUID,
    },
    AllProp: true,
  }
  multiGet.Limit = 10

  obj, err := cd.QueryAddressBook(ab[0].Path, multiGet)

  for _, vc := range obj {
    fmt.Printf("%s: %s\n", vc.Card.Get(vcard.FieldUID).Value, vc.Card.PreferredValue(vcard.FieldFormattedName))
  }

  return
}

