# discord-bot
A bridge bot for SS13&lt;->Discord communication via webhooks and world.Topic()

> [!WARNING]
> The current implementation is simplistic, unpolished, and is intended for internal use.  
> You may, however, use this code as a reference or base for your own projects, to the extent permitted by the License.

## Supported features
- Server Status Updates
- Game-to-Discord account verification and linking
- Bi-directional chat relay for Adminhelp and OOC
- Gameserver metrics recording

Each feature can be disabled via config, allowing you to run just what you need.

## Usage example
- Install `Go` https://go.dev/
- Copy `config.example.yaml` to `config.yaml` and fill in the config
- Run `go build .` from cmd/git bash
- Run the resulting executable

## Service requirements
For configuration requirements see `config.example.yaml`  
For details on discord channel permissions see [Discord Developer Docs](https://discord.com/developers/docs/topics/permissions#permissions)

<details>
<summary>Status updates module</summary>

- Discord bot must have permission
  - view the status updates channel (`VIEW_CHANNEL`)
  - send messages (`SEND_MESSAGES`)
  - attach embeds (`EMBED_LINKS`)
  - read previous messages (`READ_MESSAGE_HISTORY`)
- Channel must be isolated from all other use and be dedicated exclusively to the status updates messages from the bot.
- Status message must be accessible from the first GetMessages call to Discord API. If there are any additional messages in the channel and bot can't find its message to edit, it'll post a new one.
</details>

<details>
<summary>Account verification module</summary>

- Game database (MariaDB, schema and scripts [here](https://github.com/ZeroHubProjects/ZeroOnyx/tree/master/tools/sql)) must be running and have schema applied
  - Required tables: `discord_player`, `verification`
- Discord bot must have permission to:
  - view the verification channel (`VIEW_CHANNEL`)
  - send messages (`SEND_MESSAGES`)
  - delete messages (`MANAGE_MESSAGES`)
  - attach embeds (`EMBED_LINKS`)
  - manage channel permissions (`MANAGE_ROLES`)
  - read previous messages (`READ_MESSAGE_HISTORY`)
</details>

<details>
<summary>Chat relay module</summary>

- Webhooks access key on the server must match access key in `config.yaml`
- Discord bot must have permission to:
  - view the [Ahelp/OOC/Emotes] channel (`VIEW_CHANNEL`)
  - send messages (`SEND_MESSAGES`)
  - delete messages (`MANAGE_MESSAGES`)
  - read previous messages (`READ_MESSAGE_HISTORY`)
  - for Ahelp channel: mention @everyone and @here (`MENTION_EVERYONE`)
</details>

<details>
<summary>Metrics module</summary>

- Game database (MariaDB, schema and scripts [here](https://github.com/ZeroHubProjects/ZeroOnyx/tree/master/tools/sql)) must be running and have schema applied
  - Required table: `server_metrics`
</details>
