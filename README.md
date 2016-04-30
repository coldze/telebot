# Framework for making Telegram bots

## Two modes are available:
* Polling mode bot
* Webhook mode bot

## Examples:
* 01_simple_bot - polling bot, that polls updates, replies with 'echo' on texts and with sticker on stickers.
* 02_command_handlers - polling bot like 01_simple_bot, but has commands support - /rem, /list
* 03_webhook_bot - bot, that gets updates through web-hook, with exact functionality, as 02_command_handlers
