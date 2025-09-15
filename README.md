
## AltavinGo

**AltavinGo** is a Discord bot written in Go, designed to interact with various AI language models (LLMs) such as **Ollama** and **Groq** via their APIs.
The bot supports flexible customization, maintains chat history, stores user and channel data, and provides convenient commands for interaction and management.
You can add it to a Discord server or use it in direct messages (DMs).


## Features

* **Support for OpenAI API and alternative LLM providers (Ollama, Groq)** – seamless integration with different models for generating responses.
* **Customizable personality and behavior** – define system prompts to shape how the bot interacts.
* **Slash command support** – intuitive interaction through Discord slash commands.
* **Chat history management with autosave** – stores per-user and per-channel history on a timer.
* **Request rate limiting and cooldown system** – prevents spam and controls workload.
* **Configurable via `config.json`** – easy setup and fine-tuning.
* **Logging support** – track bot activity.
* **Works in both Discord servers and direct messages (DMs)**.


## Quick Start

1. Install Go: [https://golang.org/dl/](https://golang.org/dl/)

2. Clone the repository:

   ```sh
   git clone https://github.com/rpastuhov/AltavinGo.git
   ```

3. Navigate to the project folder:
   ```sh
   cd AltavinGo
   ```

4. Copy the example configuration:
   ```sh
   cp config.json.example config.json
   ```

5. Fill in the required parameters in `config.json` (Discord token, API keys for models, etc.).

6. Run the bot:
   ```sh
   go run main.go
   ```


## Docker Usage
Before running with Docker, create the required files:
```sh
touch config.json chats.json guilds.json bot.log
```

To run with Docker, use:
```sh
docker compose up --build
```

## License

This project is licensed under the **Apache License 2.0**.
See [LICENSE](https://github.com/rpastuhov/AltavinGo/blob/main/LICENSE) for details.
