# AltavinGo

AltavinGo is a platform for integrating and utilizing large language models (LLM) with support for OpenAI API in Discord. This project enables you to easily configure and run a chatbot based on various LLMs, such as Ollama and Groq, with flexible configuration options.

## Features

- Support for OpenAI API and alternative LLM services (Ollama, Groq).
- Customizable bot behavior and personality through system prompts.
- Slash command registration for easy bot interaction.
- Chat history logging and saving through a timer.
- User request rate limiting and cooldown system.

## Setup

1. Clone the repository and navigate to the project directory.
   ```bash
   git clone https://github.com/rpastuhov/AltavinGo.git
   cd AltavinGo
   ```

2. In the project files, you will find a sample configuration file (`config.json`) that is already set up to work with [Ollama](https://github.com/ollama/ollama). The only thing you need to do is [create a Discord bot](https://discord.com/developers/applications) and insert the token into the appropriate field.

3. You can also use the API from [Groq](https://groq.com) or other LLM providers by adjusting the `base_url` parameter.

Here is an example of the `config.json` file:
```json
{
  "tokenDiscord": "your-discord-token",
  "tokenLLM": "ollama",
  "historyTimer": 30,
  "base_url": "localhost:11434",
  "model": "llama3:8b",
  "system_prompt": "You are a helpful assistant.",
  "max_tokens": 250,
  "temperature": 0.5,
  "register_slash_commands": false,
  "maxUserRequests": 10,
  "cooldown_time": 30
}
```

### Configuration Parameters:
- **tokenDiscord**: The token used to authenticate your bot with Discord.
- **tokenLLM**: The token to connect to the LLM service (e.g., `ollama` or a token from Groq).
- **historyTimer**: The interval (in minutes) for saving chat history to a file.
- **base_url**: The base URL of your LLM provider.
- **model**: The name of the language model to use.
- **system_prompt**: Custom instructions defining the bot's behavior and personality.
- **max_tokens**: The maximum number of tokens in a single response.
- **temperature**: The sampling temperature for controlling creativity and randomness of responses.
- **register_slash_commands**: Enable or disable the registration of slash commands.
- **maxUserRequests**: The maximum number of requests a user can make before hitting the timeout.
- **cooldown_time**: The cooldown period (in seconds) before the user can interact with the bot again after reaching the request limit.

## License

This project is distributed under the Apache License 2.0. See [LICENSE](https://github.com/rpastuhov/AltavinGo/blob/main/LICENSE) for further details.
