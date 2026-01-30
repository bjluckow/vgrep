# `vgrep`: Visual Global Regex Print

Test [regular expressions](https://en.wikipedia.org/wiki/Regular_expression) in real time. Powered by [grep](https://en.wikipedia.org/wiki/Grep).
![demo](demo.png)

---

## Installation

Note that `vgrep` **assumes and depends on** a local `grep` installation.

### Option 1 — Go install

```bash
go install github.com/bjluckow/vgrep@latest
```

Make sure `$GOPATH/bin` or `$HOME/go/bin` is in your PATH.

### Option 2 — build locally

```bash
git clone https://github.com/bjluckow/vgrep
cd vgrep
go build
```

Then move the binary somewhere in your PATH:

```bash
mv vgrep ~/.local/bin/
```

---

## Usage

`vgrep` reads from stdin for Unix-style composability.

Pipe input:

```bash
cat file.txt | vgrep
```

Or pass a file:

```bash
vgrep file.txt
```

Then, type a regular expression to match lines as per `grep`.

Press **Enter** to emit results to stdout.

Press **Esc** or **Ctrl+C** to cancel.


### Flags

Run `vgrep --help` to view native flags.

`vgrep` directly wraps `grep`. You can pass raw flags to `grep` after `--`:

```bash
vgrep -d -- -i
```

---
