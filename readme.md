# add-osc-8-hyperlink

_read from stdin, find relative or absolute paths, output osc-8 `file://` hyperlinks to terminal_

[for terminal emulators that support osc-8](https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda)

#### installation

`go install go.senan.xyz/add-osc-8-hyperlink@latest`

#### example integrations

##### fish and git diff, status, log

```fish
function git
    if isatty stdout; and contains -- $argv[1] diff status log
        command git -c color.status=always -c color.ui=always $argv | add-osc-8-hyperlink
        return
    end
    command git $argv
end
```

##### fish and ripgrep

```fish
function rg
    if isatty stdout
        command rg --color=always --line-number $argv | add-osc-8-hyperlink
        return
    end
    command rg $argv
end
```

##### fish and git video

<https://user-images.githubusercontent.com/6832539/199340070-d34a6a38-8fce-49c3-a16a-32e88dad870e.mp4>
