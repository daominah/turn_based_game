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
- **Action Log**: The engine maintains a generic action log with sequence numbers (1, 2, 3, ...) and timestamps for each action. This enables replay functionality and allows players to review the full history of the duel. Games can log actions using `Duel.LogAction()`.

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

### Message Flow

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

**Scalability Note**: The code should allow scaling the server to run on multiple machines easily; the first implementation is for a single instance.

## Front end

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
- Supports join URLs: players can share a URL with `?duelId=X&playerId=Y` query parameters to automatically join a duel
- Browser URL automatically updates to the join URL after creating or joining a duel, allowing refresh without losing access
- Auto-connects to WebSocket on page load

### User Interface Layout

The UI features a **3-column layout**:

- **Left Sidebar**: Technical details and controls
  - Connection status with visual indicators for connection state, connection duration logged in browser console and server logs
  - Create duel form (hidden after creating a duel or when joining via URL)
  - Default player IDs: Alice_XXXXXX and Bob_XXXXXX (where XXXXXX is a random 6-digit number)
  - Duel information (ID, turn, current player, state, winner when ended)
  - Join URLs for each player with player-colored labels (input field and copy button on separate lines for easier use)
  - Wider width to display full duel ID
  - Join via URL: players can use URLs with `?duelId=X&playerId=Y` query parameters to automatically join

- **Center Column**: Duel board (main game area)
  - **Turn Info**: Displays at the top showing current turn number and turn owner, or "Duel Ended - Winner: [player]" when the duel ends
  - **Top Player Area**: 3x2 grid layout with colored border
    - Top-left: Deck (shows card count)
    - Top-mid: Hand (face-down cards, displayed in single row with horizontal scroll if needed)
    - Top-right: empty
    - Bot-left: Graveyard (shows card count and last played card)
    - Bot-mid: empty
    - Bot-right: Player ID and Life Points (LP on new line)
  - **Middle Play Zone**: Same height as player areas. When a card is played, it appears here for several seconds with animation, then moves to graveyard. The chosen option (Gain or Inflict) is highlighted with distinct colors
  - **Bottom Player Area**: 3x2 grid layout (same structure as top player)
    - Top-left: Player ID and Life Points (LP on new line)
    - Top-mid: empty
    - Top-right: Graveyard (shows card count and last played card)
    - Bot-left: End Turn button (styled as a card, aligned left with player info) or empty depending on the turn or duel ended
    - Bot-mid: Hand (face-up cards with action buttons, displayed in single row with horizontal scroll if needed)
    - Bot-right: Deck (shows card count)
  - Deck and Graveyard have the same height for alignment
  - Player zones are color-coded: distinct colors for each player (colors assigned server-side for consistency)
  - Turn indicator: Active player's zone has distinct background color, thicker border, and shadow effect
  - End Turn button always visible for current player (disabled when not their turn)
  - Compact design to fit the entire board without scrolling

- **Right Sidebar**: Duel log
  - Shows all actions from all players (public log, part of generic engine)
  - Each entry has a sequence number (1, 2, 3, ...) for replay ordering
  - Format: `2006-01-02T15:04:05.999: PlayerID: [Gain X] [Inflict Y]`
  - The chosen option (Gain or Inflict) is highlighted with distinct colors and bold text
  - Each row is color-coded by player (assigned consistently based on player order in duel)
  - Timestamp shows when the action occurred
  - Can be used to replay the entire duel from the beginning
  - Automatically scrolls to show the latest action

**Card Design**:
- Cards are simplified to show only two action buttons: "Gain X" and "Inflict Y"
- No card ID or other technical details displayed
- Buttons are color-coded with distinct colors for each action type
- Cards are disabled when it's not the player's turn
- Opponent's cards are shown face-down (opacity reduced with "?" indicator)
