function getBackendUrl() {
    if (typeof process !== "undefined" && process.env && process.env.BACKEND_URL) {
        return process.env.BACKEND_URL;
    }
    return "http://localhost:11995";
}

function getWebSocketUrl() {
    const httpUrl = getBackendUrl();
    return httpUrl.replace(/^http/, "ws") + "/ws";
}

const BACKEND_URL = getBackendUrl();
const WS_URL = getWebSocketUrl();

// Track last played card for animation
let lastPlayedCard = null;
let playZoneTimeout = null;

// Player colors (Blue and Purple)
const playerColors = {
    // Colors will be assigned dynamically based on player order
};

/**
 * Logs messages with a timestamp and caller location.
 * @param {...any} args - The messages to log.
 */
function log(...args) {
    const now = new Date().toISOString();
    const stack = new Error().stack.split("\n")[2] || "";
    const match = stack.match(/at\s+(.*):(\d+):(\d+)/);
    let location = "";
    if (match) {
        const file = match[1].split(/[\\/]/).pop();
        const line = match[2];
        location = `${file}:${line}`;
    }
    console.info(`[${now}] ${location}`, ...args);
}

/**
 * Calls the /api/hello endpoint and displays the result in the helloResult div.
 * @returns {Promise<void>}
 */
async function callHelloApi() {
    const helloResult = document.getElementById("helloResult");
    if (!helloResult) {
        return;
    }
    let helloAPI = BACKEND_URL + "/api/hello";
    const beginT = Date.now();
    log(`begin fetch ${helloAPI}`);
    helloResult.textContent = "Loading...";
    try {
        const res = await fetch(helloAPI);
        helloResult.textContent += await res.text();
        const duration = Date.now() - beginT;
        log(`end fetch ${helloAPI} (duration: ${duration}ms)`);
    } catch (err) {
        helloResult.textContent += "error: " + err;
        const duration = Date.now() - beginT;
        log(`error fetch ${helloAPI}, duration: ${duration}ms)`);
    }
}

// WebSocket connection
let ws = null;
let isConnecting = false;
let reconnectTimeout = null;
let currentDuelId = null;
let currentPlayerId = null;
let currentGameState = null;
let previousGameState = null;

/**
 * Connects to WebSocket server
 * Automatically called on page load and reconnects on disconnect
 */
function connectWebSocket() {
    // Prevent multiple simultaneous connection attempts
    if (isConnecting) {
        log("WebSocket connection already in progress, skipping...");
        return;
    }

    if (ws && ws.readyState === WebSocket.OPEN) {
        log("WebSocket already connected");
        return;
    }

    // Clear any pending reconnection
    if (reconnectTimeout) {
        clearTimeout(reconnectTimeout);
        reconnectTimeout = null;
    }

    isConnecting = true;
    const connectStartTime = performance.now();
    log(`Connecting to WebSocket: ${WS_URL}`);
    updateConnectionStatus("Connecting");

    try {
        ws = new WebSocket(WS_URL);

        ws.onopen = function () {
            isConnecting = false;
            const connectDuration = (performance.now() - connectStartTime).toFixed(2);
            log(`WebSocket connected in ${connectDuration}ms`);
            updateConnectionStatus("Connected", WS_URL);

            // If we have a duel/player, rejoin automatically
            if (currentDuelId && currentPlayerId) {
                log("Rejoining duel after reconnection...");
                joinDuel(currentDuelId, currentPlayerId);
            }
        };

        ws.onmessage = function (event) {
            try {
                const message = JSON.parse(event.data);
                handleServerMessage(message);
            } catch (err) {
                log("Error parsing server message:", err);
            }
        };

        ws.onerror = function (error) {
            log("WebSocket error:", error);
            updateConnectionStatus("Error - Check server connection");
            isConnecting = false;
        };

        ws.onclose = function (event) {
            isConnecting = false;
            log(`WebSocket closed (code: ${event.code}, reason: ${event.reason || 'none'})`);
            updateConnectionStatus("Disconnected - Reconnecting...");

            // Only reconnect if it wasn't a normal closure
            if (event.code !== 1000) {
                // Attempt to reconnect after 3 seconds
                reconnectTimeout = setTimeout(() => {
                    reconnectTimeout = null;
                    connectWebSocket();
                }, 3000);
            }
        };
    } catch (err) {
        isConnecting = false;
        log("Error creating WebSocket:", err);
        updateConnectionStatus("Error - Failed to create connection");
        // Retry after 3 seconds
        reconnectTimeout = setTimeout(() => {
            reconnectTimeout = null;
            connectWebSocket();
        }, 3000);
    }
}

/**
 * Handles messages from server
 */
function handleServerMessage(message) {
    log("Received server message:", message);

    switch (message.type) {
        case "state_update":
            // Detect if a card was just played by checking the last action in the log
            // Only show animation for PLAY_CARD actions, not END_TURN
            if (message.duel && message.duel.action_log && message.duel.action_log.length > 0) {
                const lastAction = message.duel.action_log[message.duel.action_log.length - 1];

                if (lastAction.action === "PLAY_CARD" && message.game_state && message.game_state.players) {
                    // Find the player who played the card and get the card from their graveyard
                    const playerId = lastAction.player_id;
                    const player = message.game_state.players[playerId];

                    if (player && player.graveyard && player.graveyard.length > 0) {
                        // Get the last card from this player's graveyard
                        const lastCard = player.graveyard[player.graveyard.length - 1];
                        lastPlayedCard = lastCard;

                        // Clear after 8 seconds
                        if (playZoneTimeout) {
                            clearTimeout(playZoneTimeout);
                        }
                        playZoneTimeout = setTimeout(() => {
                            lastPlayedCard = null;
                            if (currentGameState) {
                                updateGameUI(currentGameState);
                            }
                        }, 8000);
                    }
                } else if (lastAction.action === "END_TURN") {
                    // Clear any existing animation when turn ends
                    if (playZoneTimeout) {
                        clearTimeout(playZoneTimeout);
                        playZoneTimeout = null;
                    }
                    lastPlayedCard = null;
                }
            }

            previousGameState = currentGameState;
            currentGameState = message;

            // Handle duel ending gracefully
            try {
                updateGameUI(message);
            } catch (err) {
                log("Error updating game UI:", err);
                // Still try to show the state even if there's an error
                if (message.duel && message.duel.state === 'END') {
                    log("Duel ended - winner:", message.duel.winner);
                }
            }
            break;
        case "error":
            log("Server error:", message.error);
            // Don't use alert for errors, just log them
            console.error("Server error:", message.error);
            break;
        default:
            log("Unknown message type:", message.type);
    }
}

/**
 * Sends a message to the server
 */
function sendMessage(message) {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
        alert("WebSocket not connected. Please connect first.");
        return;
    }

    log("Sending message:", message);
    ws.send(JSON.stringify(message));
}

/**
 * Creates a new duel
 */
function createDuel() {
    const game = "CARD_GAME_BURN";
    const player1 = document.getElementById("player1Input").value || "player1";
    const player2 = document.getElementById("player2Input").value || "player2";

    const message = {
        type: "create_duel",
        game: game,
        players: [player1, player2]
    };

    sendMessage(message);
    currentPlayerId = player1; // Assume first player is this client

    // Hide Create Duel section after creating
    hideCreateDuelSection();

    // URL will be updated when we receive the state update with duel ID
}

/**
 * Hides the Create Duel section
 */
function hideCreateDuelSection() {
    const sections = document.querySelectorAll('.section');
    sections.forEach(section => {
        const h3 = section.querySelector('h3');
        if (h3 && h3.textContent === 'Create Duel') {
            section.style.display = 'none';
        }
    });
}

/**
 * Joins an existing duel
 */
function joinDuel(duelId, playerId) {
    // Allow passing parameters directly (for URL join) or reading from URL params
    if (!duelId || !playerId) {
        const urlParams = new URLSearchParams(window.location.search);
        duelId = duelId || urlParams.get('duelId');
        playerId = playerId || urlParams.get('playerId');
    }

    if (!duelId || !playerId) {
        alert("Please provide both Duel ID and Player ID");
        return;
    }

    const message = {
        type: "join_duel",
        duel_id: duelId,
        player_id: playerId
    };

    sendMessage(message);
    currentDuelId = duelId;
    currentPlayerId = playerId;
}

/**
 * Sends a play card action
 */
function playCard(cardId, option) {
    if (!currentDuelId || !currentPlayerId) {
        alert("Please create or join a duel first");
        return;
    }

    const message = {
        type: "action",
        duel_id: currentDuelId,
        player_id: currentPlayerId,
        game: "CARD_GAME_BURN",
        action: {
            card_id: cardId,
            option: option
        }
    };

    sendMessage(message);
}

/**
 * Sends an end turn action
 */
function endTurn() {
    if (!currentDuelId || !currentPlayerId) {
        alert("Please create or join a duel first");
        return;
    }

    const message = {
        type: "action",
        duel_id: currentDuelId,
        player_id: currentPlayerId,
        game: "CARD_GAME_BURN",
        action: {
            end_turn: true
        }
    };

    sendMessage(message);
}

/**
 * Updates the game UI with current state
 */
function updateGameUI(stateMessage) {
    if (!stateMessage.duel || !stateMessage.game_state) {
        return;
    }

    const duel = stateMessage.duel;
    const gameState = stateMessage.game_state;

    // Update player colors from server (source of truth)
    if (duel.player_colors) {
        Object.assign(playerColors, duel.player_colors);
    }

    // Update duel info in sidebar
    const duelInfoDiv = document.getElementById("duelInfo");
    if (duelInfoDiv && duel.id) {
        // Generate join URLs for both players
        const playerIds = Object.keys(gameState.players || {});
        let joinUrlsHTML = '';
        if (playerIds.length > 0) {
            playerIds.forEach((pid, idx) => {
                const joinUrl = generateJoinUrl(duel.id, pid);
                const playerColor = playerColors[pid] || '#000';
                const inputId = `joinUrl_${pid.replace(/[^a-zA-Z0-9]/g, '_')}`;
                joinUrlsHTML += `
                    <p style="margin-top: ${idx === 0 ? '10' : '5'}px;"><strong style="color: ${playerColor};">Join as ${pid}:</strong></p>
                    <input type="text" id="${inputId}" readonly value="${joinUrl}"
                        style="font-size: 0.7em; padding: 3px 4px; width: 100%; word-break: break-all; overflow-wrap: break-word; border: 1px solid #ccc; border-radius: 3px; background: #fff; color: #000; box-sizing: border-box; margin-bottom: 3px;"
                        onclick="this.select();">
                    <button onclick="copyToClipboard('${inputId}')"
                        style="padding: 3px 6px; font-size: 0.65em; white-space: nowrap; cursor: pointer; background: #6c757d; color: white; border: none; border-radius: 3px; display: flex; align-items: center; gap: 3px; width: 100%; justify-content: center; margin-bottom: 5px;">
                        <span style="font-size: 0.75em;">📋</span><span>Copy</span>
                    </button>`;
            });
        }

        duelInfoDiv.innerHTML = `
			<p><strong>Duel ID:</strong> ${duel.id}</p>
			<p><strong>Turn:</strong> ${duel.turn}</p>
			<p><strong>Current Player:</strong> ${duel.turn_player}</p>
			<p><strong>State:</strong> ${duel.state}</p>
			${duel.winner ? `<p><strong>Winner:</strong> ${duel.winner}</p>` : ""}
			${joinUrlsHTML}
		`;
    }

    // Update duel log display
    updateDuelLogDisplay();

    // Update duel board
    const boardContent = document.getElementById("duelBoardContent");
    if (!boardContent || !gameState.players) {
        return;
    }

    // Update player colors from server (source of truth)
    if (duel.player_colors) {
        Object.assign(playerColors, duel.player_colors);
    }

    // Determine top and bottom players
    // Top = opponent, Bottom = current player
    const playerIds = Object.keys(gameState.players);

    let topPlayerId, bottomPlayerId;

    if (currentPlayerId && gameState.players[currentPlayerId]) {
        // Current player exists, put them at bottom
        bottomPlayerId = currentPlayerId;
        topPlayerId = playerIds.find(id => id !== currentPlayerId) || playerIds[0];
    } else {
        // No current player set, use first two players (sorted)
        bottomPlayerId = playerIds[playerIds.length - 1];
        topPlayerId = playerIds[0];
    }

    const topPlayer = gameState.players[topPlayerId];
    const bottomPlayer = gameState.players[bottomPlayerId];

    const isTopTurn = duel.turn_player === topPlayerId;
    const isBottomTurn = duel.turn_player === bottomPlayerId;
    const isBottomCurrent = bottomPlayerId === currentPlayerId;

    // Build the duel board
    let boardHTML = '<div class="duel-board-content">';

    // Turn Info at the top
    let turnInfoText = `Turn ${duel.turn}`;
    if (duel.state === 'END' && duel.winner) {
        turnInfoText = `Duel Ended - Winner: ${duel.winner}`;
    } else {
        turnInfoText += ` - Current Player: ${duel.turn_player}`;
    }
    boardHTML += `
        <div class="turn-info">
            ${turnInfoText}
        </div>
    `;

    // Get player colors for borders
    const topPlayerColor = playerColors[topPlayerId] || '#007bff';
    const bottomPlayerColor = playerColors[bottomPlayerId] || '#6f42c1';

    // Top Player Area - 3x2 grid layout (symmetric)
    boardHTML += `
		<div class="player-area top ${isTopTurn ? 'turn' : ''}" style="border-color: ${topPlayerColor};">
			<div class="player-grid">
				<div class="player-grid-cell top-left">
					<div class="deck">
						<div class="deck-label">Deck</div>
						<div class="deck-count">${topPlayer.deck_size}</div>
					</div>
				</div>
				<div class="player-grid-cell top-mid">
					<div class="player-hand">
						${topPlayer.hand.map(() => '<div class="card" style="opacity: 0.3;"><div style="text-align: center; padding-top: 50px;">?</div></div>').join("")}
					</div>
				</div>
				<div class="player-grid-cell top-right"></div>
				<div class="player-grid-cell bot-left">
					<div class="graveyard">
						<div class="graveyard-label">GY: ${topPlayer.graveyard ? topPlayer.graveyard.length : 0}</div>
						${topPlayer.graveyard && topPlayer.graveyard.length > 0 ? `
							<div class="graveyard-card">
								<div>Gain: ${topPlayer.graveyard[topPlayer.graveyard.length - 1].gain}</div>
								<div>Inflict: ${topPlayer.graveyard[topPlayer.graveyard.length - 1].inflict}</div>
							</div>
						` : ''}
					</div>
				</div>
				<div class="player-grid-cell bot-mid"></div>
				<div class="player-grid-cell bot-right">
					<div class="player-info">
						<span class="player-name">${topPlayerId}${isTopTurn ? ' (Turn)' : ''}</span>
						<br>
						<span class="player-lp">LP: ${Math.floor(topPlayer.life_point)}</span>
					</div>
				</div>
			</div>
		</div>
	`;

    // Middle Play Zone - show last played card temporarily
    boardHTML += `
		<div class="play-zone ${lastPlayedCard ? 'has-card' : ''}" id="playZone">
			${lastPlayedCard ? `
				<div class="played-card">
					<p><strong>Card Played</strong></p>
					<p class="card-option ${lastPlayedCard.played_option === 'GAIN' ? 'chosen gain-option' : ''}">Gain: ${lastPlayedCard.gain}</p>
					<p class="card-option ${lastPlayedCard.played_option === 'INFLICT' ? 'chosen inflict-option' : ''}">Inflict: ${lastPlayedCard.inflict}</p>
				</div>
			` : '<p style="color: #6c757d;">Play zone: cards appear here when played</p>'}
		</div>
	`;

    // Bottom Player Area (Current Player) - 3x2 grid layout (symmetric)
    boardHTML += `
		<div class="player-area bottom ${isBottomTurn ? 'turn' : ''}" style="border-color: ${bottomPlayerColor};">
			<div class="player-grid">
				<div class="player-grid-cell top-left">
					<div class="player-info">
						<span class="player-name">${bottomPlayerId}${isBottomCurrent ? ' (You)' : ''}${isBottomTurn ? ' (Turn)' : ''}</span>
						<br>
						<span class="player-lp">LP: ${Math.floor(bottomPlayer.life_point)}</span>
					</div>
				</div>
				<div class="player-grid-cell top-mid"></div>
				<div class="player-grid-cell top-right">
					<div class="graveyard">
						<div class="graveyard-label">GY: ${bottomPlayer.graveyard ? bottomPlayer.graveyard.length : 0}</div>
						${bottomPlayer.graveyard && bottomPlayer.graveyard.length > 0 ? `
							<div class="graveyard-card">
								<div>Gain: ${bottomPlayer.graveyard[bottomPlayer.graveyard.length - 1].gain}</div>
								<div>Inflict: ${bottomPlayer.graveyard[bottomPlayer.graveyard.length - 1].inflict}</div>
							</div>
						` : ''}
					</div>
				</div>
				<div class="player-grid-cell bot-left">
					${isBottomCurrent && duel.state !== 'END' ? `
						<div class="end-turn-container">
							<button class="end-turn-card" onclick="endTurn()" ${isBottomTurn ? '' : 'disabled'}>
								<div class="end-turn-text">End Turn</div>
							</button>
						</div>
					` : ''}
				</div>
				<div class="player-grid-cell bot-mid">
					<div class="player-hand">
						${bottomPlayer.hand.map(card => `
							<div class="card ${isBottomTurn && isBottomCurrent ? '' : 'disabled'}">
								<button class="card-button gain"
									onclick="playCard('${card.unique_card_id}', 'GAIN')"
									${isBottomTurn && isBottomCurrent ? '' : 'disabled'}>
									Gain ${card.gain}
								</button>
								<button class="card-button inflict"
									onclick="playCard('${card.unique_card_id}', 'INFLICT')"
									${isBottomTurn && isBottomCurrent ? '' : 'disabled'}>
									Inflict ${card.inflict}
								</button>
							</div>
						`).join("")}
					</div>
				</div>
				<div class="player-grid-cell bot-right">
					<div class="deck">
						<div class="deck-label">Deck</div>
						<div class="deck-count">${bottomPlayer.deck_size}</div>
					</div>
				</div>
			</div>
		</div>
	`;

    boardHTML += '</div>';
    boardContent.innerHTML = boardHTML;

    currentDuelId = duel.id;

    // Hide Create Duel section if duel is active
    if (duel.id) {
        hideCreateDuelSection();
    }

    // Update browser URL to reflect join URL for current player
    if (duel.id && currentPlayerId) {
        const newUrl = `${window.location.pathname}?duelId=${encodeURIComponent(duel.id)}&playerId=${encodeURIComponent(currentPlayerId)}`;
        window.history.replaceState({}, '', newUrl);
    }
}

/**
 * Copies text to clipboard
 */
function copyToClipboard(inputId) {
    const input = document.getElementById(inputId);
    if (input) {
        input.select();
        input.setSelectionRange(0, 99999); // For mobile devices
        try {
            document.execCommand('copy');
            log("Copied to clipboard:", input.value);
            // Visual feedback
            const button = input.nextElementSibling;
            if (button) {
                const originalText = button.textContent;
                button.textContent = "Copied!";
                button.style.background = "#28a745";
                setTimeout(() => {
                    button.textContent = originalText;
                    button.style.background = "#007bff";
                }, 1000);
            }
        } catch (err) {
            log("Failed to copy:", err);
        }
    }
}

/**
 * Updates the duel log display from server state
 */
function updateDuelLogDisplay() {
    const logDiv = document.getElementById("duelLog");
    if (!logDiv) return;

    // Get action log from duel (generic, not game-specific)
    if (!currentGameState || !currentGameState.duel || !currentGameState.duel.action_log) {
        logDiv.innerHTML = '<p class="empty-message">Log will appear here</p>';
        return;
    }

    const actionLog = currentGameState.duel.action_log;

    // Update player colors from server (source of truth)
    if (currentGameState.duel && currentGameState.duel.player_colors) {
        Object.assign(playerColors, currentGameState.duel.player_colors);
    }

    if (actionLog.length === 0) {
        logDiv.innerHTML = '<p class="empty-message">Log will appear here</p>';
        return;
    }

    let logHTML = '';
    actionLog.forEach(entry => {
        const playerColor = playerColors[entry.player_id] || '#000';

        // Format log entry based on action type
        let actionText = '';
        if (entry.action === 'PLAY_CARD' && entry.data) {
            const option = entry.data.option || '';
            const gain = entry.data.gain || 0;
            const inflict = entry.data.inflict || 0;
            // Use color instead of checkmark: green for GAIN, red for INFLICT
            const gainColor = option === 'GAIN' ? 'color: #28a745; font-weight: bold;' : '';
            const inflictColor = option === 'INFLICT' ? 'color: #dc3545; font-weight: bold;' : '';
            actionText = `[<span style="${gainColor}">Gain ${gain}</span>] [<span style="${inflictColor}">Inflict ${inflict}</span>]`;
        } else {
            actionText = `Action: ${entry.action}`;
        }

        logHTML += `
            <div class="duel-log-entry" style="color: ${playerColor}; margin-bottom: 5px;">
                ${entry.timestamp}: ${entry.player_id}: ${actionText}
            </div>
        `;
    });

    logDiv.innerHTML = logHTML;
    // Scroll to bottom
    logDiv.scrollTop = logDiv.scrollHeight;
}

/**
 * Generates a join URL for the current duel
 */
function generateJoinUrl(duelId, playerId) {
    const baseUrl = window.location.origin + window.location.pathname;
    return `${baseUrl}?duelId=${encodeURIComponent(duelId)}&playerId=${encodeURIComponent(playerId)}`;
}

/**
 * Updates connection status display
 * Shows connection status with server address when connected
 */
function updateConnectionStatus(status, serverAddress) {
    const statusDiv = document.getElementById("connectionStatus");
    if (statusDiv) {
        if (status === "Connected" && serverAddress) {
            statusDiv.textContent = `Connected to ${serverAddress}`;
            statusDiv.className = "status connected";
            statusDiv.style.display = "block";
        } else {
            statusDiv.textContent = "Status: " + status;
            statusDiv.className = "status " + status.toLowerCase();
            statusDiv.style.display = "block";
        }
    }
}

/**
 * Generates a random 6-digit number string
 */
function generateRandomSuffix() {
    let result = '';
    for (let i = 0; i < 6; i++) {
        result += Math.floor(Math.random() * 10);
    }
    return result;
}

/**
 * Initializes event handlers and logs window load events.
 * @returns {void}
 */
window.onload = function () {
    log("begin window.onload");

    // Check for join URL parameters
    const urlParams = new URLSearchParams(window.location.search);
    const duelIdFromUrl = urlParams.get('duelId');
    const playerIdFromUrl = urlParams.get('playerId');

    // Set default player IDs with random suffixes
    const player1Input = document.getElementById("player1Input");
    const player2Input = document.getElementById("player2Input");
    if (player1Input) {
        player1Input.value = "Alice_" + generateRandomSuffix();
    }
    if (player2Input) {
        player2Input.value = "Bob_" + generateRandomSuffix();
    }

    // If join URL parameters exist, hide Create Duel section and auto-join
    if (duelIdFromUrl && playerIdFromUrl) {
        // Hide Create Duel section for join users
        hideCreateDuelSection();

        // Auto-join after WebSocket is connected
        const checkConnection = setInterval(() => {
            if (ws && ws.readyState === WebSocket.OPEN) {
                clearInterval(checkConnection);
                joinDuel(duelIdFromUrl, playerIdFromUrl);
            }
        }, 100);

        // Timeout after 5 seconds
        setTimeout(() => clearInterval(checkConnection), 5000);
    }

    // Connect WebSocket on load
    connectWebSocket();

    // Setup button handlers
    const helloBtn = document.getElementById('helloBtn');
    const helloResult = document.getElementById('helloResult');
    if (helloBtn && helloResult) {
        helloBtn.onclick = callHelloApi;
    }

    const createDuelBtn = document.getElementById("createDuelBtn");
    if (createDuelBtn) {
        createDuelBtn.onclick = createDuel;
    }


    setTimeout(function () {
        log("end window.onload");
    }, 1000);
};

/**
 * Toggles sidebar visibility
 */
function toggleSidebar(side) {
    const sidebar = document.getElementById(side === 'left' ? 'leftSidebar' : 'rightSidebar');
    const showButton = document.getElementById(side === 'left' ? 'showLeftSidebar' : 'showRightSidebar');

    if (sidebar.classList.contains('hidden')) {
        sidebar.classList.remove('hidden');
        showButton.style.display = 'none';
    } else {
        sidebar.classList.add('hidden');
        showButton.style.display = 'block';
    }
}
