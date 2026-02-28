# Discord Archival Tool

Discord Archive Tool is meant to be a tool for Archival. The goal is to make it easy to archive everything on a Discord server. Great emphasis is put on archiving which will mean that this tool must ARCHIVE FUCKING EVERYTHING!

# Requirements
* Have gcc installed and set `CGO_ENABLED=1` in the Enviromental variables.


Updates:
~~I am using `modernc.org/sqlite` instead of `go-sqlite3` to avoid downloading the C compiler and enabling `CGO_ENABLED=1`. I hope to go back to using `go-sqlite3` as I do not wish to comprimise performance.~~
You need gcc to set `CGO_ENABLED=1` for this to work.

This repository is licensed with the [MIT](https://mit-license.org/) license.