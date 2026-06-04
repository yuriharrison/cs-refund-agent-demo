# Architecture Principles

- **AI-Native & Conversational:** The UI design philosophy centers around natural interaction.
- **Observable AI:** AI actions and function calls must be exposed to the frontend as system debug messages.
- **Snapshot Testing:** Unit tests must use a snapshot pattern to mock OpenAI API calls after the first run.
- **Separation of Concerns:** Frontend uses generated API clients, backend handles agent logic and DB interactions.
