package store

import (
  "encoding/json"
  "strings"

  "github.com/agnivade/levenshtein"
  "github.com/emersion/go-vcard"
  "github.com/tidwall/buntdb"
)

type Store struct {
  db          *buntdb.DB
}

func Open(path string) (*Store, error) {
  var err error
  s := new(Store)

  s.db, err = buntdb.Open(path)
  if err != nil {
    return nil, err
  }

  return s, nil
}

func (s *Store) Close() () {
  s.db.Close()
}

func (s *Store) Upsert(vcs []*vcard.Card) (error) {
  err := s.db.Update(func(tx *buntdb.Tx) error {
    for _, vc := range vcs {
      mvc, err := json.Marshal(vc)
      if err != nil {
        return err
      }
      tx.Set(vc.Get(vcard.FieldUID).Value, string(mvc), nil)
    }
    return nil
  })
  return err
}

func (s *Store) FindBy(key string, val string) ([]vcard.Card, error) {
  var vcards []vcard.Card

  err := s.db.View(func(tx *buntdb.Tx) error {
    return tx.Ascend("", func(k, v string) bool {
      var vc vcard.Card
      err := json.Unmarshal([]byte(v), &vc)
      if err != nil {
        return true
      }

      vcv := vc.PreferredValue(key)

      vcTokens := strings.Split(vcv, " ")
      valTokens := strings.Split(val, " ")

      distances := make([]int, len(valTokens))
      for x := 0; x < len(valTokens); x++ {
        distances[x] = 100
      }

      for i := 0; i < len(valTokens); i++ {
        for j := 0; j < len(vcTokens); j++ {
          distance := levenshtein.ComputeDistance(
            vcTokens[j],
            valTokens[i],
          )
          if distance < distances[i] {
            distances[i] = distance
          }
        }

      }

      totalDistance := 0
      for x := 0; x < len(valTokens); x++ {
        totalDistance += distances[x]
      }
      if totalDistance <= (len(distances) * 2) {
        vcards = append(vcards, vc)
      }

      return true
    })
  })

  return vcards, err
}

func (s *Store) FindByFn(fn func(vc *vcard.Card) bool) ([]vcard.Card, error) {
  var vcards []vcard.Card

  err := s.db.View(func(tx *buntdb.Tx) error {
    return tx.Ascend("", func(k, v string) bool {
      var vc vcard.Card
      err := json.Unmarshal([]byte(v), &vc)
      if err != nil {
        return true
      }

      if add := fn(&vc); add == true {
        vcards = append(vcards, vc)
      }

      return true
    })
  })

  return vcards, err
}

