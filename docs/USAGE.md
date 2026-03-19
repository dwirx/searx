# Usage Guide

## Check Version

```bash
search --version
search version
```

`search version` shows both CLI and Lightpanda versions.

## Search Web Results

Default engine:

```bash
search "golang generics"
```

Select engine:

```bash
search -e ddg "linux kernel scheduler"
search -e brave "zero trust architecture"
search -e google "go 1.25 release notes"
search -e mojeek "privacy search engine"
```

Searx with custom instance:

```bash
search -e searx -i https://searx.be "open source intelligence"
```

## Hacker News Mode

```bash
search -e hn -hn top
search -e hn -hn best
search -e hn -hn ask
```

Supported HN categories: `top`, `new`, `best`, `ask`, `show`, `job`.

## Read and Save Articles

Read article content:

```bash
search -read "https://go.dev/blog/go1.22"
```

Read and save to Markdown:

```bash
search -read "https://www.nytimes.com/2026/03/17/world/middleeast/iran-war-israel-middle-east-recap.html" -save
```

Force Lightpanda:

```bash
search -read "https://example.com/article" -panda -save
```

Force archive.today prefix:

```bash
search -read "https://example.com/article" -archive -save
```

## Lightpanda Management

```bash
search setup   # install/check/update to latest if needed
search update  # force update check now
```

## Output Files

When `-save` is used, output is written to the current directory as:

`<sanitized-title>.md`
