package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/araddon/dateparse"
	"github.com/eliukblau/pixterm/pkg/ansimage"

	"github.com/emersion/go-vcard"
	"github.com/mrusme/addrb/dav"
	"github.com/mrusme/addrb/store"
)


func main() () {
  var err        error
  var username   string
  var password   string
  var endpoint   string
  var addrbDb    string
  var addrbTmpl  string

  var refresh    bool
  var birthdays  bool
  var lookupAttr string
  var outputJson bool

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
  flag.StringVar(
    &addrbTmpl,
    "template",
    os.Getenv("ADDRB_TEMPLATE"),
    "addrb template file",
  )

  flag.BoolVar(
    &refresh,
    "r",
    false,
    "Refresh local vcard database",
  )
  flag.BoolVar(
    &birthdays,
    "birthdays",
    false,
    "List contacts that have their birthday today",
  )
  flag.StringVar(
    &lookupAttr,
    "l",
    vcard.FieldFormattedName,
    "Lookup attribute",
  )
  flag.BoolVar(
    &outputJson,
    "j",
    false,
    "Output JSON",
  )

  flag.Parse()

  args := flag.Args()

  if len(args) == 0 &&
     refresh == false &&
     birthdays == false {
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
    if err != nil {
      fmt.Printf("%s\n", err)
      os.Exit(1)
    }

    err = cd.RefreshAddressBooks()
    if err != nil {
      fmt.Printf("%s\n", err)
      os.Exit(1)
    }

    paths := cd.GetAddressBookPaths()
    vcs := cd.GetVcardsInAddressBook(paths[0])

    err = db.Upsert(vcs)
    if err != nil {
      fmt.Printf("%s\n", err)
      os.Exit(1)
    }
  }

  if len(args) == 0 &&
  birthdays == false {
    os.Exit(0)
  }

  var t *template.Template
  if len(addrbTmpl) > 0 && outputJson == false {
    t = template.Must(template.New("addrb").Funcs( template.FuncMap{
      "RenderPhoto": func(photo string) string {
        return RenderPhoto(photo)
      },
      "RenderAddress": func(address string) string {
        return RenderAddress(address)
      },
      "RenderBirthdate": func(dt string) string {
        return RenderBirthdate(dt)
      },
    }).ParseFiles(addrbTmpl))
  }

  var foundVcs []vcard.Card
  var foundBdays []time.Time
  var today time.Time = time.Now()

  if birthdays == true {

    foundVcs, err = db.FindByFn(func(vc *vcard.Card) bool {
      dt, errr := dateparse.ParseAny(vc.PreferredValue(vcard.FieldBirthday))
      if errr == nil {
        if dt.Month() == today.Month() &&
           dt.Day()   == today.Day()    {
          foundBdays = append(foundBdays, dt)
          return true
        }
      }

      return false
    })
  } else {
    foundVcs, err = db.FindBy(lookupAttr, lookupVal)
  }
  if err != nil {
    fmt.Printf("%s\n", err)
    os.Exit(1)
  }

  for idx, vc := range foundVcs {
    photo := vc.PreferredValue(vcard.FieldPhoto)
    photoRender := RenderPhoto(photo)

    if outputJson == true {
      b, err := json.MarshalIndent(vc, "", "  ")
      if err != nil {
        fmt.Printf("%s\n", err)
        os.Exit(1)
      }

      fmt.Printf(string(b))
      os.Exit(0)
    } else {
      if len(addrbTmpl) > 0 {
        err := t.ExecuteTemplate(os.Stdout, path.Base(addrbTmpl), vc)
        if err != nil {
          fmt.Printf("%s\n", err)
          os.Exit(1)
        }
      } else {
        if birthdays == true {
          fmt.Printf(
            "%s (%d)\n",
            vc.PreferredValue(vcard.FieldFormattedName),
            (today.Year() - foundBdays[idx].Year()),
          )
        } else {
          fmt.Printf(
            "\n%s\n%s\n----------------------------------------\nBirthday:  %s\nTel.:      %s\nEmail:     %s\n\nAddress:\n%s\n\n",
            photoRender,
            vc.PreferredValue(vcard.FieldFormattedName),
            RenderBirthdate(vc.PreferredValue(vcard.FieldBirthday)),
            vc.PreferredValue(vcard.FieldTelephone),
            vc.PreferredValue(vcard.FieldEmail),
            RenderAddress(vc.PreferredValue(vcard.FieldAddress)),
          )
        }
      }
    }
  }

  os.Exit(0)
}

func RenderPhoto(photo string) (string) {
  photoRender := ""
  if len(photo) > 0 {
    reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(photo))
    m, _, err := image.Decode(reader)
    if err == nil {
      pix, err := ansimage.NewScaledFromImage(
        m,
        20,
        20,
        color.Transparent,
        ansimage.ScaleModeResize,
        ansimage.NoDithering,
      )
      if err == nil {
        photoRender = pix.RenderExt(false, false)
      }
    }
  }
  return photoRender
}

func RenderAddress(address string) (string) {
  addr := strings.Split(address, ";")

  switch(len(addr)) {
  case 0:
    return ""
  case 1:
    return addr[0]
  case 7:
    var str = ""
    if len(addr[0]) > 0 {
      str = fmt.Sprintf("%s%s\n", str, addr[0])
    }
    if len(addr[1]) > 0 {
      str = fmt.Sprintf("%s%s\n", str, addr[1])
    }
    if len(addr[2]) > 0 {
      str = fmt.Sprintf("%s%s\n", str, addr[2])
    }
    if len(addr[5]) > 0 && len(addr[3]) > 0 {
      str = fmt.Sprintf("%s%s %s\n", str, addr[5], addr[3])
    }
    if len(addr[4]) > 0 {
      str = fmt.Sprintf("%s%s\n", str, addr[4])
    }
    if len(addr[6]) > 0 {
      str = fmt.Sprintf("%s%s\n", str, addr[6])
    }
    return str
  }

  return ""
}

func RenderBirthdate(dt string) string {
  t, err := dateparse.ParseAny(dt)
  if err != nil {
    return dt
  }

  return t.Format("02 Jan 2006")
}

