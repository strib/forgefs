[![Go Reference](https://pkg.go.dev/badge/github.com/strib/forgefs.svg)](https://pkg.go.dev/github.com/strib/forgefs) [![Go Report Card](https://goreportcard.com/badge/github.com/strib/forgefs)](https://goreportcard.com/report/github.com/strib/forgefs)

# forgefs -- The Keyforge Filesystem

Finally!  A filesystem that lets you browse all Keyforge cards and all
the decks you have registered at
[decksofkeyforge.com](https://decksofkeyforge.com).  forgefs is a
filesystem that runs on Ubuntu and MacOS that shows up like a thumb
drive, letting you browse both cards and decks from the command line
or in your file broswer.

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
the amazing [SkyJedi](https://mas.to/@SkyJedi):

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
* Count your deck with "perfect" stats (amber control >= 10, expected
  amber >= 20, creature control >= 10, artifact control >= 1.5, and
  efficiency >= 10):
  ![Filter on perfect stats](https://user-images.githubusercontent.com/8516691/216799194-482ccc62-72be-4fa5-8ffe-5df552154658.png)

