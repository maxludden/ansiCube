# Renders a single screenshot showing the entire 256-color palette.
# Run from the repository root after building ./ansicube:
#   vhs docs/tapes/palette.tape

Output docs/assets/screenshots/palette.gif

Require ./ansicube

Set Shell "bash"
Set FontSize 12
Set LineHeight 1.0
Set Width 1100
Set Height 1300
Set Padding 16
Set Theme "Catppuccin Mocha"

Hide
Type "./ansicube"
Enter
Sleep 1.5s
Show

Sleep 500ms
Screenshot docs/assets/screenshots/palette.png

Hide
Ctrl+C
