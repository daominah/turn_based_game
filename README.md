# Generic turn-based game

This repo contains a generic turn-based game engine,
backend written in Go, frontend written in JavaScript.

## Back end

### Generic Game Engine

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
