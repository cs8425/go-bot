
const fetchReq = (url, option = {}, successFn, errFn) => {
	const req = fetch(url, option).then(function (res) {
		if (!res.ok) throw res.text();
		return res.json();
	});
	if (typeof successFn === 'function') req.then(successFn);
	if (typeof errFn === 'function') req.catch(async function (err) {
		if (typeof err.then === 'function') {
			err = await err;
		}
		errFn(err);
	});
	return req;
}

export { fetchReq };

const dumpJson = (el, data, fileName) => {
	// console.log('[dump]', data);
	// const el = dummyDlEl.current;
	let json = JSON.stringify(data);
	let blob = new Blob([json], { type: "octet/stream" });
	let url = window.URL.createObjectURL(blob);
	el.href = url;
	el.download = fileName || 'config.json';
	el.click();
	setTimeout(() => { window.URL.revokeObjectURL(url); }, 30 * 1000);
};

export { dumpJson };

async function PBKDF2_deriveKey(pwd, iter = 4096, salt = null) {
	const keyMaterial = await crypto.subtle.importKey(
		"raw",
		new TextEncoder().encode(pwd),
		"PBKDF2",
		false,
		["deriveBits", "deriveKey"]
	);

	return await crypto.subtle.deriveKey(
		{
			"name": "PBKDF2",
			salt: salt,
			"iterations": iter,
			"hash": "SHA-256"
		},
		keyMaterial,
		{ "name": "AES-GCM", "length": 256 },
		true,
		["encrypt", "decrypt"]
	);
}
function buf2hex(buffer) { // buffer is an ArrayBuffer
	// https://stackoverflow.com/questions/40031688/javascript-arraybuffer-to-hex
	return [...new Uint8Array(buffer)].map(x => x.toString(16).padStart(2, '0')).join('');
}
function hex2buf(hex) { // return Uint8Array, get ArrayBuffer by `.buffer`
	// https://gist.github.com/don/871170d88cf6b9007f7663fdbc23fe09
	// https://blog.csdn.net/sinat_36728518/article/details/117132147
	return new Uint8Array(hex.match(/[\da-f]{2}/gi).map((h) => parseInt(h, 16)));
}
async function sha256Sum(dataStr) {
	// https://developer.mozilla.org/en-US/docs/Web/API/SubtleCrypto/digest
	const msgUint8 = new TextEncoder().encode(dataStr);                 // encode as (utf-8) Uint8Array
	const hashBuffer = await crypto.subtle.digest('SHA-256', msgUint8); // hash the message
	return buf2hex(hashBuffer);
}
async function sha256SumBuf(data) {
	const hashBuffer = await crypto.subtle.digest('SHA-256', data);
	return buf2hex(hashBuffer);
}

const cryptoApi = {
	isEncrypt(obj) {
		return obj.enc && obj.iv && obj.salt && obj.mac;
	},
	async decrypt(obj, pwd, iter = 131072) { // iter = 2^17
		const { enc, iv, salt, mac } = obj;
		const cyphertext = hex2buf(enc).buffer;

		// check mac of cyphertext
		const ck = await sha256SumBuf(cyphertext);
		if (ck !== mac) {
			throw 'cropped data or wrong password';
		}
		const key = await PBKDF2_deriveKey(pwd, iter, hex2buf(salt));
		const cleartext = await crypto.subtle.decrypt({ name: 'AES-GCM', tagLength: 32, iv: hex2buf(iv) }, key, cyphertext);
		const json = JSON.parse(new TextDecoder().decode(cleartext));
		return json;
	},
	async encrypt(data, pwd, iter = 131072) { // iter = 2^17
		const text = JSON.stringify(data, '');
		const salt = crypto.getRandomValues(new Uint8Array(16));
		const key = await PBKDF2_deriveKey(pwd, iter, salt);
		const iv = await crypto.getRandomValues(new Uint8Array(32)) // IV must be the same length (in bits) as the key
		const cyphertext = await crypto.subtle.encrypt({ name: 'AES-GCM', tagLength: 32, iv }, key, new TextEncoder().encode(text))

		// TODO: base64?
		const obj = {
			enc: buf2hex(cyphertext),
			iv: buf2hex(iv.buffer),
			salt: buf2hex(salt.buffer),
			mac: await sha256SumBuf(cyphertext), // TODO
		};
		return obj;
	},
	buf2hex,
	hex2buf,
	sha256Sum,
	sha256SumBuf,
};

export { cryptoApi };
