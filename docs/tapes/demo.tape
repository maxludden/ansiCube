# Renders an animated demo plus screenshots of scrolling and the quit prompt.
# Run from the repository root after building ./ansicube:
#   vhs docs/tapes/demo.tape

Output docs/assets/screenshots/demo.gif

Require ./ansicube

Set Shell "bash"
Set FontSize 16
Set Width 1200
Set Height 720
Set Padding 16
Set Theme "Catppuccin Mocha"
Set TypingSpeed 120ms

Hide
Type "./ansicube"
Enter
Sleep 1.5s
Show

Sleep 1s
Screenshot docs/assets/screenshots/top.png

# Scroll line by line, then a page at a time.
Down@150ms 6
Sleep 500ms
PageDown
Sleep 1s
Screenshot docs/assets/screenshots/cube.png

PageDown 10
Sleep 1s
Screenshot docs/assets/screenshots/grayscale.png

PageUp 10
Sleep 800ms

# q asks for confirmation; n cancels.
Type "q"
Sleep 1s
Screenshot docs/assets/screenshots/quit-prompt.png
Type "n"
Sleep 800ms

# Confirm this time.
Type "q"
Sleep 600ms
Type "y"
Sleep 500ms
