# Generic turn-based game

This repo contains a generic turn-based game engine,
backend written in Go, frontend written in JavaScript.

## Back end

### Generic turn-based game engine

A highly generic implementation of a turn-based game engine, containing only the
common logic required by all turn-based games.

- Supports any number of **players** (default is 2).
- The **duel** always has a state. Three main states are:
  - BEGIN: The duel has just been initialized. Some automatic actions are performed,
    such as tossing a coin to determine who plays first, drawing cards, or placing
    chess pieces in their starting positions. No player can perform actions in this state.
  - END: The duel has ended, either because someone has won or, rarely, due to a draw.
  - RUNNING: The duel is in progress. Only one player can perform valid actions at a time.
    After each action, the game state changes, and the engine determines which player
    can act next and what actions are available, according to the game logic.
- Valid actions are determined by the current game state.
- Logging is recommended: all actions and significant state changes should be logged
  so players can review them during the duel or reconstruct a replay later.

### Pluggable games logic

#### Burn card game

Very simple card game to demonstrate the engine.

- At the beginning of the duel, toss a coin to determine who plays first.
- Each player has 8000 LP and draws 5 cards. Then start the 1st turn.
- At the start of each turn, except the 1st turn, the turn player draws 1 card.
- All cards have similar effects:
  Activate 1 of the following effects:
  - Gain x LP.
  - Inflict y damage to the opponent.
    (LP values are randomly generated: 0 < x < 1000, 0 < y < 3000, divisible by 100)
- The duel ends when one player's LP reaches 0 (or less),
  or when a player needs to draw but the deck is empty.
- During a turn, only the turn player can perform actions. Actions are:
  - Play a card from hand.
  - End turn.

#### Real game

TODO.

## Architecture

### Message Flow (Slack-like pattern)

The system follows a three-stage message flow pattern similar to Slack's architecture:

1. **Message In (Ingress)**: When a player performs an action (e.g., plays a card, ends turn, creates a duel), the frontend sends the action via WebSocket to the backend. This provides real-time, low-latency communication.

2. **Persist**: The backend immediately persists the action and resulting game state changes to storage (currently in-memory, can be extended to database). This ensures durability and allows for replay/reconstruction of game history.

3. **Fanout**: After persistence, the backend pushes the updated game state to all connected clients (players in the duel) via WebSocket. This ensures all players see the same state simultaneously without polling.

**Note**: For simplicity, this server uses WebSocket for all communication. HTTP REST API could be used for non-real-time operations (e.g., creating duels, querying history), but WebSocket is used throughout for consistency and simplicity.

This pattern ensures:

- **Low Latency**: WebSocket bidirectional communication provides faster response times
- **Simplicity**: Single communication protocol for all operations
- **Reliability**: State is persisted before distribution, so no data is lost
- **Consistency**: All players receive the same state update
- **Scalability**: The persist-then-fanout pattern allows for future horizontal scaling (e.g., using message queues or pub/sub systems)

### Front end

Static files are served from the [web](web) directory.

**Vanilla JavaScript implementation** (no frameworks required):

- Simple and lightweight - vanilla JS is powerful enough for this use case
- Uses native `WebSocket` API for all communication (actions and state updates)
- Minimal dependencies, easy to understand and maintain

The frontend:

- Maintains a WebSocket connection for all game communication
- Sends all player actions through the WebSocket connection (creating duels, playing cards, ending turns, etc.)
- Receives game state updates through the same WebSocket connection
- Updates the UI reactively when state changes arrive

The code should allow scaling the server to run on multiple machines easily;
the first implementation is for a single instance.
