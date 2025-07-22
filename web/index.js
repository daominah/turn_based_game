function getBackendUrl() {
	if (typeof process !== "undefined" && process.env && process.env.BACKEND_URL) {
		return process.env.BACKEND_URL;
	}
	return "http://localhost:11995";
}

const BACKEND_URL = getBackendUrl();

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

/**
 * Initializes event handlers and logs window load events.
 * @returns {void}
 */
window.onload = function () {
	log("begin window.onload");
	const helloBtn = document.getElementById('helloBtn');
	const helloResult = document.getElementById('helloResult');
	if (helloBtn && helloResult) {
		helloBtn.onclick = callHelloApi;
	}
	setTimeout(function () {
		log("end window.onload");
	}, 1000);
};
