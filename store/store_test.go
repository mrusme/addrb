package store

import (
  "testing"
  "path/filepath"
  "io"
  "os"

  "github.com/emersion/go-vcard"
)

func loadTestCards(t *testing.T, fileName string) []*vcard.Card {
  t.Helper()

  f, err := os.Open(filepath.Join("testdata", fileName))
  if err != nil {
    t.Fatalf("Failed to open %s: %v", fileName, err)
  }
  defer f.Close()

  cards := make([]*vcard.Card, 0)

  dec := vcard.NewDecoder(f)
  for {
    card, err := dec.Decode()
    if err == io.EOF {
      break
    } else if err != nil {
      t.Fatalf("Failed to load card: %v", err)
    }
    cards = append(cards, &card)
  }
  t.Logf("Loaded %d cards from %s", len(cards), fileName)
  return cards

}

func TestOpenNewAddReopenFind(t *testing.T) {

  // Create new store, tempdir will auto-delete
  storePath := filepath.Join(t.TempDir(), "addrb.db")

  // Load some cards from testdata
  cards := loadTestCards(t, "4-cards.vcf")

  // Create new store
  store, err := Open(storePath)
  if err != nil {
    t.Fatalf("Failed to open: %v", err)
  }
  defer store.Close()

  // Add cards
  if err := store.Upsert(cards); err != nil {
    t.Fatalf("Failed to add cards: %v", err)
  }

  store.Close()

  // Reopen store
  store, err = Open(storePath)
  if err != nil {
    t.Fatalf("Failed to reopen: %v", err)
  }
  defer store.Close()

  foundCards, err := store.FindBy(vcard.FieldFormattedName, "Doe")
  if err != nil {
    t.Fatalf("Failed to find: %v", err)
  }

  t.Logf("Found %d cards: ", len(foundCards))
  for _, card := range foundCards {
    t.Logf(" - %s", card.PreferredValue(vcard.FieldFormattedName))
  }

  if len(foundCards) < 2 {
    t.Fatalf("Expected at least 2 cards, got %d", len(foundCards))
  }
}
