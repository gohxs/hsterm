termu (Work in progress)
=============

Terminal Utilities

My own implementation of a tool similar to Readline
(code with some inspiration on the chzyer/readline project)

Readline, Ansi parser, ansi writer for windows

* Display handler - readbuffer output for highlighters/completers
* Autocomplete handler - similar to the one found in golang.org/x/crypto/ssh/terminal

TODO
----------

* [ ] 256/16m to 16 color translator for windows or check terminal support for
  more colors
* [ ] spf13/cobra example to make it like netsh windows
* [ ] Native multi line suport that allows multiline completion
* [ ] Fish alike fuzzy search of history
  * [X] Created an example of how it would be implemented, earlier versions
* [X] Char cursor memory instead of line output
* Create and rearrange windows Term
  * [X] Writer wrapper
  * [X] Movement Esc
  * [X] Colors
  * [X] Cleaners \x1B[J and \x1B[K
  * [X] Cursor visibility (but blinks alot)
  * [X] Issue on last column on windows
  * [ ] Windows suports VT by passing necessary flags to output, check
    the possibility on input
    [setconsolemode](https://docs.microsoft.com/en-us/windows/console/setconsolemode)
  * [ ] Check windows version or capability of the VT flag
* Handle unicode/double width chars (Partial)
  * [X] Print unicode
  * [X] String manipulation
  * [ ] Cursor manipulation runeWidth
* [ ] Prepare completer maybe (readline)
* [ ] History search (readline)
* [ ] Alter display to preview tabs (fish)
* [X] Project rename, hsterm was *not* a good name
* [X] Show a progress bar and prompt at same time
