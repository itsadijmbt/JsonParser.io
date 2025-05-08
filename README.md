# JsonParser.io
Go TUI JSON-Tree Parser

A powerful terminal UI tool for parsing, visualizing, and exploring JSON data directly in your console. Built in Go with low-level JSON parsing, a custom AST-style tree builder, and interactive UI using Bubbletea & Lipgloss.

🚀 Features

Low-Level JSON Parsing: Custom streaming lexer + decoder with zero-copy token handling and minimal allocations for blazing-fast performance.

AST-Style Node Builder: Recursively converts maps and arrays into a navigable Node tree in memory.

Dynamic TUI View: Interactive terminal interface powered by Bubbletea & Lipgloss, featuring:

Resizable viewport on window resize

Adjustable indent width via ←/→ keys

Smooth tick-driven, animated line reveal

Thick branded borders and color schemes

Precision Glyph Rendering: Accurate ├ vs. └ connectors and perfectly aligned branches at any depth.

Keyboard Controls:

j/down & k/up to scroll

←/→ to change indent

q/esc to quit

📥 Download (Windows Only)

Grab the latest Windows executable from our Releases page.
# Download and run
Invoke-WebRequest -Uri "https://github.com/itsadijmbt/JsonParser.io/releases/tag/JSP1.1"
.\\json-tree-parser.exe -file config.json

Currently, pre-built binaries are provided only for Windows. Linux and macOS users should build from source.


