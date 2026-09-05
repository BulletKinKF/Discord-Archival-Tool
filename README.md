This allows you to resume the archival process of a Discord server. The servers are archived as Text only.

### Goals
* Allow medias to be organized with SHA-256
* Convert database to other formats (HTML, JSON)
* Implement Server <--> Client functionality. This should allow the user to use multiple accounts for archival.
* Automatic habitual archival of updated messages.

***
# Discord Archival Tool

Discord Archive Tool is meant to be a tool for Archival. The goal is to make it easy to archive everything on a Discord server. Great emphasis is put on archiving which will mean that this tool must ARCHIVE EVERYTHING!

# Requirements
* Have gcc installed and set `CGO_ENABLED=1` in the Enviromental variables.

You need gcc to set `CGO_ENABLED=1` for this to build.

This repository is licensed with the [GPL 3.0](https://www.gnu.org/licenses/gpl-3.0.en.html/) license.

***

## Alternative Tools:
[DiscordChatExporter](https://github.com/Tyrrrz/DiscordChatExporter) - Most popular tool to export Discord chat.

[Discord-History-Tracker](https://github.com/chylex/Discord-History-Tracker) - Desktop app & browser script that downloads Discord chat history.

[Discrub](https://github.com/prathercc/discrub-ext) - Brower extension that saves Discord chat. It is written in TypeScript, I consider this the most slowest but most convenient tool.

[Discord-DL](https://github.com/Yakabuff/discord-dl) - An old Discord archival tool written in Golang. It seeks to acheive the same goals as this repo.