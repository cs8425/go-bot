
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
