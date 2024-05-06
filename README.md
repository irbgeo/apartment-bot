// TODO: update

# Apartment-bot

The Apartment-bot is a comprehensive bot designed to facilitate apartment hunting. The bot seamlessly integrates three distinct services to streamline the process of finding and disseminating apartment listings.

## 1. server

The server service serves as the backbone of the bot, actively querying apartment aggregators to compile and maintain a robust database of available apartments. It constantly monitors the relevance of the data and efficiently sends out apartment listings based on user-defined filters.

## 2. Client

The Client service functions as the user interface, enabling interactions between the bot and the client. Users can create personalized filters, submit apartment preferences, and receive tailored listings. This service ensures a user-friendly experience in the apartment search process.

### Parameters

| Configuration Parameter                  | Type          | Environment Variable                            | Default Value                                  | Description                                                     |
| ---------------------------------------- | ------------- | ----------------------------------------------- | ---------------------------------------------- | --------------------------------------------------------------- |
| serverURL                                | string        | server_URL                                      | localhost:9000                                 | URL of the apartment server                                     |
| MessageURL                               | string        | MESSAGE_URL                                     | localhost:9001                                 | URL for messaging service                                       |
| TelegramBotSecret                        | string        | TELEGRAM_BOT_SECRET                             | 6327864323:AAHPqArVfe6fzgMZfoaHWciLmuQmbaQUSpc | Secret key for the Telegram bot                                 |
| TelegramBotSendPeriod                    | time.Duration | TELEGRAM_BOT_SEND_PERIOD                        | 10s                                            | Period for sending messages to Telegram users                   |
| TelegramBotMaxCountSendMessagesPerPeriod | int64         | TELEGRAM_BOT_MAX_COUNT_SEND_MESSAGES_PER_PERIOD | 10                                             | Maximum count of messages to send per period                    |
| TelegramBotAdminUsername                 | string        | TELEGRAM_BOT_ADMIN_USERNAME                     | rent_apartment_georgia_bot_admin               | Username of the Telegram bot admin                              |
| TelegramBotDisabledParameters            | []string      | TELEGRAM_BOT_DISABLED_PARAMS                    |                                                | List of parameters for disabling                                |
| FirstCities                              | []string      | FIRST_CITIES                                    | Tbilisi,Batumi                                 | List of initial cities displayed in the filter setup            |
| AuthToken                                | string        | AUTH_TOKEN                                      | test                                           | Security token for authentication (replace with a secure token) |

### TelegramBotDisabledParameters

List of actions to be turned off for configuration. Parameters for disabling:

- change_owner_type_action - owner type configuration.
- change_type_action - apartment type configuration - for sale or for rent.

## 3. Message

The Message service facilitates communication by allowing the bot to send informative messages to all its clients. This feature ensures timely updates, announcements, and other relevant information is efficiently communicated to the user base.

### Key Features

**Aggregator Integration** server service actively queries apartment aggregators for the latest listings.

**Custom Filters** Clients can create and modify filters to receive personalized apartment recommendations.

**Real-time Updates** The server constantly updates the apartment database, ensuring listing relevance.

**User Interaction** The Client service facilitates seamless communication between the bot and users.

**Broadcast Messages** The Message service enables the bot to send informative messages to all clients.

Experience the convenience of Apartment-bot, your dedicated assistant in the quest for the perfect apartment!
