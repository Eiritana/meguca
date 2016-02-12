/*
 Initialises the SockJS server
 */

const config = require('../../config'),
	fs = require('fs'),
	{Client} = require('../websockets'),
	util = require('./util'),
	winston = require('winston')

const sockJs = require('sockjs').createServer({
	prefix: config.SOCKET_PATH,
	jsessionid: false,
	log: sockjs_log,
	websocket: config.USE_WEBSOCKETS
})

/**
 * Forward sockJS warnings and errors to winston
 * @param {string} sev
 * @param {string} message
 */
function sockjs_log(sev, message) {
	if (sev === 'info')
		winston.verbose(message)
	else if (sev === 'error')
		winston.error(message)
}

/**
 * Create Client() for each websocket connection
 */
sockJs.on('connection', socket => {
	const client = new Client(socket, socket.ip)
	socket.on('data', data => client.onMessage(data))
	socket.on('close', () => client.onClose())
});

/**
 * Attach SockJS handler to HTTP server
 * @param {http.Server} server
 */
export function start(server) {
	server.on('upgrade', (req, res) => res.end())
	sockJs.installHandlers(server)
}