
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
