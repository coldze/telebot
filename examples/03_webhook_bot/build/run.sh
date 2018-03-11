#!/bin/bash
#set -e

export BOT_TOKEN=346931272:AAEeQJcfl566XU9o5fuHlP88Zb8aSRi83bw
# 202857756:AAF841dy1EUA_OF36FQ2CUCIqJ6GuQtrnkk
export BOT_SSL_PUBLIC=/etc/letsencrypt/live/coldzedev.ddns.net/fullchain.pem
export BOT_SSL_PRIVATE=/etc/letsencrypt/live/coldzedev.ddns.net/privkey.pem
export BOT_UPDATE_CALLBACK_URL=https://coldzedev.ddns.net:8443/telebot/callback
export BOT_HTTPS_LISTEN_PORT=8443
./03_webhook_bot

