const fs = require('fs');
const esbuild = require('esbuild');

let copy = function (srcDir, dstDir) {
	var results = [];
	var list = fs.readdirSync(srcDir);
	var src, dst;
	list.forEach(function (file) {
		src = srcDir + '/' + file;
		dst = dstDir + '/' + file;
		//console.log(src);
		var stat = fs.statSync(src);
		if (stat && stat.isDirectory()) {
			try {
				console.log('creating dir: ' + dst);
				fs.mkdirSync(dst);
			} catch (e) {
				console.log('directory already exists: ' + dst);
			}
			results = results.concat(copy(src, dst));
		} else {
			try {
				console.log('copying file: ' + dst);
				//fs.createReadStream(src).pipe(fs.createWriteStream(dst));
				fs.writeFileSync(dst, fs.readFileSync(src));
			} catch (e) {
				console.log('could\'t copy file: ' + dst);
			}
			results.push(src);
		}
	});
	return results;
}

let reactOnResolvePlugin = {
	name: 'react2preact',
	setup(build) {
		// let path = require('path')

		// Redirect all 'react' to "./public/images/"
		build.onResolve({ filter: /^react$/ }, args => {
			return { path: require.resolve('preact/compat') }
		});
		build.onResolve({ filter: /^react-dom$/ }, args => {
			return { path: require.resolve('preact/compat') }
		});
	},
}

copy('./public', '../www');

console.log('start build...');
esbuild.build({
	entryPoints: ['src/main.js'],
	bundle: true,
	sourcemap: false,
	sourcesContent: true,
	target: [
		'es2015',
	],
	outfile: '../www/js/main.js',
	loader: {
		'.js': 'jsx',
	},
	jsxFactory: 'h',
	jsxFragment: 'Fragment',
	plugins: [
		reactOnResolvePlugin,
	],
}).then(()=>{
	console.log('build end!');
}).catch(() => process.exit(1));


