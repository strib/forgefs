[![Go Reference](https://pkg.go.dev/badge/github.com/strib/forgefs.svg)](https://pkg.go.dev/github.com/strib/forgefs) [![Go Report Card](https://goreportcard.com/badge/github.com/strib/forgefs)](https://goreportcard.com/report/github.com/strib/forgefs)

# forgefs -- The Keyforge Filesystem

Finally!  A filesystem that lets you browse all
(Keyforge)[https://keyforging.com/) cards and all the decks you have
registered at [decksofkeyforge.com](https://decksofkeyforge.com).
forgefs is a filesystem that runs on Ubuntu and MacOS that shows up
like a thumb drive, letting you browse both cards and decks from the
command line or in your file broswer.

## Usage

Here is an example of browsing the filesystem on Linux.  In this
example, we show:

1. Listing all Keyforge cards from the command line, then showing the
   card data and image for one card.
2. Doing the same thing, from the GUI file browser.
3. Listing all your decksofkeyforge decks, then showing the deck data
   for one of them.
4. Generate a PDF of all the cards in the deck on the fly using a
   standard Linux tool.

![Using forgefs in Linux](https://user-images.githubusercontent.com/8516691/216798897-2cd8fd29-07cd-410f-bf80-7513facddf2c.gif)

You can also see a decklist image for any of your decks, courtesy of
the amazing [SkyJedi](https://github.com/SkyJedi/):

![Seeing your deck image in Linux](https://user-images.githubusercontent.com/8516691/216798930-21879d31-3be4-40f8-a5a1-ccfc0c48343f.gif)

### Filtering decks

One of the coolest things you can do is filter your decks by different
decksofkeyforge statistics, or by what houses or sets it has, right
from the command line.

All you need to do is navigate into a _virtual directory_ in your
`my-decks` directory. This isn't a real directory that will show up in
`ls` or in your file browser listing; it only exists when you go into
it or try to `ls` it directly.  Like magic!

The name for this virtual directory describes the kind of filter you
want.  There are a bunch of stats you can filter on, all calculated by
the awesome folks at
[decksofkeyforge.com](https://decksofkeyforge.com).

* `a`: amber control
* `e`: expected amber
* `r`: artifact control
* `c`: creature control
* `f`: efficiency
* `d`: disruption
* `sas`: overall SAS score
* `aerc`: AERC score
* `expansion` or `set`: the acronym of the Keyforge set of the deck (e.g., MM)
* `house`: matches one of the houses in the deck

You can choose one of those stats, followed by an `=` and either the
exact number you want to match, or a _range_.  Ranges are one or two
numbers combined with a `:`.  The minimum number goes to the left of
the `:`, and the maximum number goes to the right.  For example:

* Amber control minimum of 10: `10:`
* Creature control maximum of 5: `:5`
* SAS between 80 and 90 (inclusive): `80:90`

What's more, you can combine these stat filters using boolean logic
and parentheses. The possible boolean operators are:

* And: `,` or `+`
* Or: `^`

Examples, assuming you have navigated into your `my-decks` directory:

* Count all your decks with SAS between 80 and 90 (inclusive):
  ![Filter on SAS](https://user-images.githubusercontent.com/8516691/216799191-ba5cc2b8-4b1c-47ac-9cf0-6d452dee5269.png)
* Count your decks with "perfect" stats (amber control >= 10, expected
  amber >= 20, creature control >= 10, artifact control >= 1.5, and
  efficiency >= 10):
  ![Filter on perfect stats](https://user-images.githubusercontent.com/8516691/216799194-482ccc62-72be-4fa5-8ffe-5df552154658.png)
* Count your decks with "perfect" stats _or_ a SAS of at least 80:
  ![Filter on perfect stats or SAS of 80](https://user-images.githubusercontent.com/8516691/216799199-43c100fd-f013-40f4-8042-98453c68a61f.png)
* Count your Mass Mutation decks with an AERC of at 68:
  ![Filter on AERC and set](https://user-images.githubusercontent.com/8516691/216799202-c4c003a6-a149-4f29-901f-162d05c1c1e3.png)
* Pick a random deck out of your MM decks with a SAS of at least 68:
  ![Random deck](https://user-images.githubusercontent.com/8516691/216799204-9a43abec-b050-4a26-a756-5bc1377cf692.png)

Note that the above examples are just counting directory entries
using a standard `wc` command line tool.  Each of those directory
entries is a full deck directory, where you can access cards and
decklist images as well.  The only limit on what you can do with that
is your imagination!

## Build/Install

forgefs is written in [Go](https://go.dev/).  Once you install and
configure Go, it is super easy to install forgefs.  This will download
the repository and install it in your `$GOPATH/bin` directory:

```sh
go get github.com/strib/forgefs
go install github.com/strib/forgefs/forgefs
```

Alternatively, we will soon provide a pre-built package for
Debian/Ubuntu.

## Configuration/Run:

The forgefs config file lives by default at
`$HOME/.forgefs_config.json` (though you can specify an alternate
config file location on the command line to the `forgefs` binary).
Before you can run it, you need to generate a decksofkeyforge API key
for yourself; you can do that
[here](https://decksofkeyforge.com/about/sellers-and-devs).  Once you
have your key, create a config file at `$HOME/.forgefs_config.json`
that looks something like this:

```json
  {
    "dok_api_key": "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
  }
```

where "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX" is your real
decksofkeyforge API key.

By default, the mountpoint for forgefs is `$HOME/ffs`, though that is
also configurable on the command line or in the config file (with a
"mountpoint" key).  That directory must exist and be empty before you
run forgefs.  forgefs also caches data on your file system, by default
in directory `$HOME/.local/share/forgefs`.  You also need to make that
directory.

```sh
mkdir -p $HOME/ffs
mkdir -p $HOME/.local/share/forgefs/forgefs_images
```

After that, you're ready to run it!  You can run `forgefs` with no
command line and starting browsing.

### Debian/Ubuntu

If you've installed our `.deb` package, you can run it with the
`run_forgefs` command, which starts up forgefs using `systemd` so it
always runs in the background.  It also creates the appropriate
directories for you.

## Credit

* [decksofkeyforge.com](https://decksofkeyforge.com) is awesome,
  consider supporting it on
  [Patreon](https://www.patreon.com/decksofkeyforge).  It assigns a
  bunch of statistics to your decks and is also a great deck
  collection manager, in addition to other things.
* SkyJedi is an amazing Keyforge community member, consider support
  him as well on [Patreon](https://www.patreon.com/SkyJedi).
