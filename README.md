# Framework for making Telegram bots (Go)

## Two modes are available:
* Polling mode bot
* Webhook mode bot

## Examples:
* 01_simple_bot - polling bot, that polls updates, replies with 'echo' on texts and with sticker on stickers.
* 02_command_handlers - polling bot like 01_simple_bot, but has commands support - /rem, /list.
* 03_webhook_bot - bot, that gets updates through web-hook, with exact functionality, as 02_command_handlers.
* 04_inline_keyboard - bot, that has a new command - /inline, will respond with inline-keyboard.
* 05_upload_photo - bot, that uploads provided image (resends, if already uploaded) in response to command /test.
