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

* [ ] Lower level Key handler instead of operating in autocompleter
* [ ] PRI: Create the Virtual Key rune based on unicode PUA
  * [X] keys are mapped, in testing
* [ ] Instead of readline, we can do a Scanner
* [ ] Package organization
  * [X] Merge termu/term with termu/term/termutils into one
  * [ ] NewStdoutWriter should be renamed or restructured to NewTerminalWriter?
* [ ] 256/16m to 16 color translator for windows or check terminal support for
  more colors
* [ ] spf13/cobra example to make it like netsh windows
* [ ] Native multi line support that allows multiline completion
* [ ] Fish alike fuzzy search of history
  * [X] Created an example of how it would be implemented, earlier version
* [X] Char cursor memory instead of line output
* WINDOWS Create and rearrange Term
  * [X] Writer wrapper
  * [X] Movement Esc
  * [X] Colors
  * [X] Cleaners \x1B[J and \x1B[K
  * [X] Cursor visibility (but blinks alot)
  * [X] Issue on last column on windows
  * [X] Windows suports VT by passing necessary flags to output, check
    the possibility on input
    [setconsolemode](https://docs.microsoft.com/en-us/windows/console/setconsolemode)
  * [X] Check windows version or capability of the VT flag
  * [ ] Support Shift tab
* Handle unicode/double width chars (Partial)
  * [X] Print unicode
  * [X] String manipulation
  * [ ] Cursor manipulation runeWidth
* [ ] Prepare completer maybe (readline)
* [X] History search (readline)
  * kind of since it is working like shell who needs history back search?
* [X] Project rename, hsterm was *not* a good name
* [X] Show a progress bar and prompt at same time
* [X] ansi.Reader is now ansi.Scanner
