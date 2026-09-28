# Themes

Press **S** on the title screen to open Settings. Use left/right to switch
between color mode and theme, up/down to preview, Space to save, and Esc to
cancel.

Typeit adapts Charmbracelet Huh's built-in palettes to its Bubble Tea screens:

| Preset | Color modes |
| --- | --- |
| Charm (initial selection) | Indigo, fuchsia, and green; light and dark variants |
| Dracula | Upstream dark palette in both modes |
| Catppuccin | Mocha in dark mode, Latte in light mode |
| Base 16 | Terminal ANSI colors in both modes |
| Default | Minimal terminal-native styling based on Huh's unstyled ThemeBase |

Bubble Tea and Lip Gloss do not provide a global theme switch. These are
adapted JSON palettes, not Huh form widgets or a new runtime dependency.
Huh itself uses Charm as its default styled theme.

The source is [Huh theme.go at ffb6a97](https://github.com/charmbracelet/huh/blob/ffb6a97195b4a031925a31cc70c62fc1d309e69d/theme.go).
Catppuccin colors come from [catppuccin/go](https://github.com/catppuccin/go).
Titles use the source title color, typed text and success use the selected-option
color, errors use the error color, and the cursor uses the input-cursor color.
Typeit supplies application backgrounds and maps metrics, warnings, and typing
feedback to those colors. Charm uses a light foreground on its dark background
and a dark foreground on cream for readability. Default adds ANSI success/error
feedback to the base styles. Theme licenses are in `assets/themes/licenses/`
and included in the generated third-party notices.

The previous GitType presets have been replaced. A saved ID that is no longer
available falls back to Charm; the replacement is persisted when settings are
saved. The `default` ID now selects the minimal Default preset.

Custom palettes retain their existing JSON format and are not overwritten:

- `~/.typeit/custom-theme.json` supplies the Custom preset. New files start
  with Charm colors.
- `~/.typeit/themes/*.json` adds named presets or overrides matching IDs.
- Missing custom color roles inherit from Charm.

`TYPEIT_DATA_DIR` relocates these files along with other application data.
