<table>
  <tr>
    <td> <img src="https://user-images.githubusercontent.com/4693125/103142905-398b8280-471d-11eb-91f6-31be10bc215c.png"  alt="1" height = 300px ></td>
    <td><img src="https://user-images.githubusercontent.com/4693125/103143012-2da0c000-471f-11eb-9d73-f8a2184c4743.jpg" alt="3" height = 300px></td>
   </tr> 
</table>

## HiLink-SMS-Resender

Periodically resends SMS from HiLink modem to your Telegram private channel

## tl;dr

```
docker run -d -e HILINK_API_URL='192.168.8.1' \
-e HILINK_USER='admin' \
-e HILINK_PASS='admin' \
-e TG_SMS_CHAN_ID='-1001405916609' \
-e TG_BOT_TKN='1130291944:AAEs_Q9dqVk55KtVZ_StKn06wMxP0fKG9VQ' \
--name hlsms wmw9/hilink-sms-resender && docker logs hlsms -f
```

Replace environment variables with your tokens
## Prerequisites

- GSM USB Modem with HiLink firmware. Working and tested on Huawei e8372h
- SIM-card that can receive SMS. Mobile data (LTE/3G) is not needed.
- usb_modeswitch installed. Your GSM USB modem should be connect in Ethernet mode
- Private Telegram channel for SMS and Telegram Bot
- Raspberry Pi (optional)

## Getting started
There are a few ways to get started.

1. Create your own Telegram private channel and get Channel ID by forwarding any message to [@myidbot](https://t.me/myidbot)
2. Create your own bot and get bot token using [@BotFather](https://t.me/botfather)
3. Add your bot to channel as administrator and give him "Post messages" permission.

### Compile from source
You'll need golang installed

```
git clone https://github.com/wMw9/hilink-sms-resender.git
cd hilink-sms-resender
```
Edit and set your environment variables by adding these lines to ~/.bashrc
```
vi ~/.bashrc
# add these lines at the end of ~/.basrc file
export HILINK_API_URL="192.168.8.1"
export HILINK_USER="admin"
export HILINK_PASS="admin"
export TG_SMS_CHAN_ID="-1241505916610" # Your Telegram Channel ID from @myidbot
export TG_BOT_TKN="12345678901:FFEs_Q9dqVkd5KtfZ_ztKnf6wM1P0gaKGhV1" # Your bot token from @botfather
```
5. Update ENVs by running `exec bash` or `. ~/.bashrc`
6. Check if your ENVs are set by `export` or `echo $HILINK_API_URL`
7. Compile from source `go build *.go -o hilink-sms-resender`
8. Make sure your modem is connected properly, up and running `ping 192.168.8.1` or whatever your IP of modem is
9. You are good to go. Run `./hilink-sms-resender`, you should see message saying `[*] HiLink SMS Resender started.`

### docker-compose

1. `git clone https://github.com/wMw9/hilink-sms-resender.git && cd hilink-sms-resender`
2. `docker-compose up --build -d && docker-compose logs -f`

```
version: '3'
services:
  hlsms:
    image: wmw9/hilink-sms-resender
    #build: .
    restart: on-failure
    networks:
      - backend
    env_file:
      - ./.env
networks:
  backend:
    driver: bridge
```

## Motivation

I've been using Apple devices for years now, they have great ecosystem. If you have iPhone and MacBook connected using one iCloud account, all your incoming SMS also popup in macOS via notification center. So you don't really need to pickup your phone to read that activation code your bank just sent you to confirm payment. This is really comfy. Also I use multiple SIM-cards and phones for security reasons like 2-factor authentication. Since I recently build myself a new Ryzen PC and now using Windows 10, I miss that feature. Luckily, I had old Huawei e8372 LTE USB modem lying around and this little project was born.

## Author

* **Ivan Smyshlyaev** - [instagram](https://instagram.com/wmw)

