# modsquad

**Batch-export tracker modules** (`.xm`, `.mod`, `.s3m`, etc.) to WAV, FLAC or MP3  
Built in Go, using the native `xmp` player plus `flac` and `lame` encoders.

---

## Features

- **Single-file or recursive** export (`-recursive`), preserving input directory tree  
- **Skip existing** exports—won’t overwrite files already in the output folder  
- **Safe interruption**: cleans up in-flight temp files on CTRL-C  
- **Default settings**:  
  - 44.1 kHz, 16 bit stereo via `xmp`  
  - FLAC compression via `flac -o`  
  - MP3 VBR ~128 kbps via `lame --vbr-new -V 6`  

---

## Requirements

- [xmp](https://xmp.sourceforge.net/) (native tracker player)  
- [flac](https://xiph.org/flac/) (for FLAC encoding) _(if `-format flac`)_  
- [lame](http://lame.sourceforge.net/) (for MP3 encoding) _(if `-format mp3`)_

---

## Installation

```bash
git clone https://github.com/verkaro/modsquad.git
cd modsquad
go build -o modsquad main.go
````

---

## Usage

```bash
# Single file to MP3 (default)
./modsquad -out exported -format mp3 song.xm

# Batch-export entire folder to FLAC, preserving tree
./modsquad -out exported -format flac -recursive path/to/modules
```

**Flags:**

* `-out` — Output root directory (creates it if needed)
* `-format` — `wav` | `flac` | `mp3`  (default `mp3`)
* `-recursive` — Recurse into input directories

---

## Contributing

Bug reports and PRs welcome! Please adhere to Go formatting and include tests where applicable.

---

## License

MIT License

---

## Credits

* Initial utility specified and iteratively implemented with help from [ChatGPT](https://openai.com).
* Conversion logic powered by `xmp`, `flac`, and `lame`.

