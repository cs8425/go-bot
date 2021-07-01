const http = require('http');
const esbuild = require('esbuild');

const config = {
	internalPort: 8000,
	devPort: 8008,
	proxyHostname: '127.0.0.1',
	proxyPort: 8001,
};

let reactOnResolvePlugin = {
	name: 'react2preact',
	setup(build) {
		// let path = require('path')

		// Redirect all 'react' to "./public/images/"
		build.onResolve({ filter: /^react$/ }, args => {
			return {
				path: require.resolve('preact/compat'),
				external: false,
			}
		});
		build.onResolve({ filter: /^react-dom$/ }, args => {
			return {
				path: require.resolve('preact/compat'),
				external: false,
			}
		});
	},
}

esbuild.serve({
	servedir: 'public',
	port: config.internalPort,
}, {
	entryPoints: ['src/main.js'],
	bundle: true,
	// minify: true,
	// pure: ['console.log'],
	sourcemap: true,
	sourcesContent: true,
	target: [
		'es2015',
	],
	outfile: 'public/js/main.js',
	loader: {
		'.js': 'jsx',
	},
	jsxFactory: 'h',
	jsxFragment: 'Fragment',
	plugins: [
		reactOnResolvePlugin,
	],
}).then(async (server) => {
	console.log('[dev]internal server start at', server.host, server.port);
	console.log('[dev]server start at', server.host, config.devPort);

	// Then start a proxy server on port 3000
	http.createServer((req, res) => {
		const options = {
			hostname: server.host,
			port: server.port,
			path: req.url,
			method: req.method,
			headers: req.headers,
		}

		const url = new URL(req.url, `http://${req.headers.host}`)
		if (url.pathname.match(/^\/api\/.*$/)) {
			options.hostname = config.proxyHostname;
			options.port = config.proxyPort;
		}

		// Forward each incoming request to esbuild
		const proxyReq = http.request(options, proxyRes => {
			// forward the response from esbuild to the client
			res.writeHead(proxyRes.statusCode, proxyRes.headers);
			proxyRes.pipe(res, { end: true });
		});
		proxyReq.on('error', (err) => {
			console.log('[req]err', req.url, err);
		});

		// Forward the body of the request to esbuild
		req.pipe(proxyReq, { end: true });
	}).listen(config.devPort);

	await server.wait;
	// Call "stop" on the web server when you're done
	// server.stop()
}).catch((err) => {
	console.log('[dev]err', err);
	process.exit(1)
});