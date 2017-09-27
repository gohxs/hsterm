hsterm (Work in progress)
=============

My own implementation of a tool similar to Readline
(code with some inspiration on the chzyer/readline project)

Readline, Ansi parser, ansi translator for windows

* Display handler - readbuffer output for highlighters/completers
* Autocomplete handler - similar to the one found in golang.org/x/crypto/ssh/terminal

TODO
----------

* [ ] 256/16m to 16 color translator for windows or check terminal support for
  more colors
* [ ] spf13/cobra example to make it like netsh windows
* [ ] Native multi line suport that allows multiline completion
* [ ] Fish alike fuzzy search of history
* [X] Char cursor memory instead of line output
* Create and rearrange windows Term
  * [X] Writer wrapper
  * [X] Movement Esc
  * [X] Colors
  * [X] Cleaners \x1B[J and \x1B[K
  * [X] Cursor visibility (but blinks alot)
  * [ ] Issue on last column on windows
* Handle unicode/double width chars (Partial)
  * [X] Print unicode
  * [X] String manipulation
  * [ ] Cursor manipulation runeWidth
* [ ] Prepare completer maybe (readline)
* [ ] History search (readline)
* [ ] Alter display to preview tabs (fish)

[X] Show a progress bar and prompt at same time
