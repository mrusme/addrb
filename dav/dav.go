package dav

import (
  "fmt"

  "github.com/emersion/go-vcard"
  "github.com/emersion/go-webdav"
  "github.com/emersion/go-webdav/carddav"
)

type DAV struct {
  httpClient      webdav.HTTPClient
  cdClient        *carddav.Client

  endpoint        string
  username        string
  password        string

  addrbookHomeSet string
  addrbooks       []carddav.AddressBook

  objects         map[string][]carddav.AddressObject
}

func New(endpoint, username, password string) (*DAV, error) {
  var err error

  dav := new(DAV)
  dav.objects = make(map[string][]carddav.AddressObject)

  dav.endpoint = endpoint
  dav.username = username
  dav.password = password

  dav.httpClient = webdav.HTTPClientWithBasicAuth(nil, dav.username, dav.password)
  dav.cdClient, err = carddav.NewClient(dav.httpClient, dav.endpoint)
  if err != nil {
    return nil, err
  }

  dav.addrbookHomeSet, err =
    dav.cdClient.FindAddressBookHomeSet(fmt.Sprintf("principals/%s", dav.username))
  if err != nil {
    return dav, err
  }

  dav.addrbooks, err = dav.cdClient.FindAddressBooks(dav.addrbookHomeSet)

  return dav, nil
}

func (dav *DAV) GetAddressBookPaths() ([]string) {
  var paths []string

  for _, ab := range dav.addrbooks {
    paths = append(paths, ab.Path)
  }

  return paths
}

func (dav *DAV) RefreshAddressBooks() (error) {
  for _, ab := range dav.addrbooks {
    err := dav.RefreshAddressBook(ab.Path)
    if err != nil {
      return err
    }
  }
  return nil
}

func (dav *DAV) RefreshAddressBook(path string) (error) {
  var err error
  query := new(carddav.AddressBookQuery)
  query.DataRequest = carddav.AddressDataRequest{
    Props: []string{
      vcard.FieldUID,
    },
    AllProp: true,
  }
  // query.Limit = 10

  dav.objects[path], err = dav.cdClient.QueryAddressBook(path, query)
  if err != nil {
    return err
  }

  return nil
}

func (dav *DAV) GetVcardsInAddressBook(path string) ([]*vcard.Card) {
  var cards []*vcard.Card

  if objs, ok := dav.objects[path]; ok {
    for i := 0; i < len(objs); i++ {
      cards = append(cards, &objs[i].Card)
    }
  }

  return cards
}

