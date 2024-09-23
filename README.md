# discord-bot
A bridge bot for SS13&lt;->Discord communication via webhooks and world.Topic()

> [!WARNING]
> The current implementation is simplistic, unpolished, and is intended for internal use.  
> You may, however, use this code as a reference or base for your own projects, to the extent permitted by the License.

## Supported features
- Server Status Updates  
- OOC chat to Discord channel bridge

Each feature can be disabled via config, allowing you to run just what you need.

## Usage example
Make sure you gave `Go` installed, see https://go.dev/  
Make a copy of `config.example.yaml` and rename it to `config.yaml`  
Edit your `config.yaml` file and fill in the values  
Open cmd/shell and navigate to this folder  
Build by running the command `go build .` in cmd/shell  
Run the produced executable which you'll find in this folder  

## Service requirements
For configuration requirements see `config.example.yaml`

### Status updates module
- The bot must have permission to view the channel, send messages to it and attach embeds.
- The channel must be isolated from all other use and be dedicated exclusively to the status updates messages from the bot.
- The status message must be accessible from the first GetMessages call to Discord API. If there are any additional messages in the channel and bot can't find its message to edit, it'll post a new one. Pagination and search are not implemented and are not currently planned to be supported.
