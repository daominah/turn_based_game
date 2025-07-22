# Generic turn-based game

This repo contains a generic turn-based game engine,
backend written in Go, frontend written in JavaScript.

## Back end

### Generic turn-based game engine

A highly generic implementation of a turn-based game engine, containing only the
common logic required by all turn-based games.

- Supports any number of players (default is 2).
- The duel always has a state. Three main states are:
  - BEGIN: The duel has just been initialized. Some automatic actions are performed,
    such as tossing a coin to determine who plays first, drawing cards, or placing
    chess pieces in their starting positions. No player can perform actions in this state.
  - END: The duel has ended, either because someone has won or, rarely, due to a draw.
  - OPEN: The duel is in progress. Only one player can perform valid actions at a time.
    After each action, the game state changes, and the engine determines which player
    can act next and what actions are available, according to the game logic.
- Valid actions are determined by the current game state.
- Logging is recommended: all actions and significant state changes should be logged
  so players can review them during the duel or reconstruct a replay later.

### Plugable games logic

#### Burn card game

Very simple card game to demonstrate the engine.

- At the beginning of the duel, toss a coin to determine who plays first.
- Each player has 8000 LP, draws 5 cards. Then start the 1st turn.
- At the start of each turn, except the 1st turn, the turn player draws 1 card.
- All cards have the similar effect:
  Activate 1 of the following effects:
  - Gain x LP.
  - Inflict y damage to the opponent.
    (the LP value is randomly generated, 0 < x < 1000, 0 < y < 3000, divisible by 100)
- The duel ends when one player LP reaches 0.
- During a turn, only the turn player can do actions. Actions are:
  - Play a card from hand.
  - End turn.

#### Real game

TODO.

## Front end

Serve directory [web](web).

Valina JavaScript and HTML.

Call backend API to send actions.

Game state is pushed from backend to frontend by WebSocket,
the code should allow to scale the server to run multiple machines easily,
the first implementation is for single instance.
