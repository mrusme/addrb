addrb
-----

![addrb](addrb.png)

`addrb`, the command line **addr**ess **b**ook.

## Build

```sh
go build .
```

## Run

Either export all necessary variables to your ENV or set them as command line
flags:

```sh
export CARDDAV_USERNAME='...'
export CARDDAV_PASSWORD='...'
export CARDDAV_ENDPOINT='...'
export ADDRB_DB='...'
```

If you're using [Baïkal](https://github.com/sabre-io/Baikal) for example, you
would export something like this as `CARDDAV_ENDPOINT`:

```sh
export CARDDAV_ENDPOINT='https://my.baik.al/dav.php/'
```

The `ADDRB_DB` is the local contacts database in order to not need to contact
the CardDAV for every lookup. You might set it to something like this:

```sh
export ADDRB_DB=~/.cache/addrb.db
```

When `addrb` is launched for the first time, it requires the `-r` flag to
refresh the contacts and sync them locally: 

```sh
addrb -r john doe
```

This command will connect to the CardDAV server, sync all address books/contacts
locally and perform a lookup for *john doe*. It will display you the contact(s) 
if any was found.

Find more flags and info with `addrb --help`.


## Templating

You can customize the regular output using templating. The template can either
be passed using the `--template <file>` flag or by exporting `ADDRB_TEMPLATE` 
in the in the environment.

The templating format is the [Go standard `text/template`][1] format. The
following special functions are available:

- `RenderPhoto` for rendering a base64 string as image (usually the contact 
  `PHOTO`)
- `RenderAddress` for rendering a contact address

Available property names that are available can be found by displaying a 
contact as JSON (using the `-j` flag). E.g. "FN" is the full name, which can be
retrieved using the `.PreferredValue` method:

```tmpl
{{ .PreferredValue "FN" }}
```

An example template can be found [here][2].


## FAQ

- Q: Does `addrb` write/modify any contact information?
  A: Nope, so far it's read-only and does not support updating vCards.
- Q: Can I use it with my local address book?
  A: Nope, as of right now `addrb` only supports CardDAV servers to sync with.
- Q: Does it support HTTP Digest auth?
  A: Nope, only HTTP Basic auth.


[1]: https://pkg.go.dev/text/template
[2]: example.tmpl

