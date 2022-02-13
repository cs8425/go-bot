
import { h, Fragment, Component, render } from 'preact';
import { useState, useEffect, useContext } from 'preact/hooks';

import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';
import { DataList } from './comp.jsx';

const header = (
	<TableRow>
		<TableCell key='act'>op</TableCell>
		<TableCell
			key='tag'
			style={{ minWidth: 250 }}
		>
			Node ID
		</TableCell>
		<TableCell
			key='time'
			align='right'
			style={{ minWidth: 150 }}
		>
			Time
		</TableCell>
		<TableCell
			key='rtt'
			align='right'
			style={{ minWidth: 150 }}
		>
			RTT
		</TableCell>
	</TableRow>
);

const formatRTT = (val) => {
	switch (true) {
		case val < 1000:
			return `${val} us`;
		case val < 1000 * 1000:
			return `${(val / 1000.0).toFixed(3)} ns`;
	}
	return `${(val / 1000.0 / 1000.0).toFixed(3)} ms`;
};

const renderRow = (v, idx) => {
	return (
		<TableRow hover role="checkbox" tabIndex={-1} key={idx}>
			<TableCell key='op'>-</TableCell>
			<TableCell key='tag'>{v.tag}</TableCell>
			<TableCell key='time' align='right'>{(new Date(v.up)).toLocaleString()}</TableCell>
			<TableCell key='rtt' align='right'>{formatRTT(v.rtt)}</TableCell>
		</TableRow>
	);
}


function NodePanel(props) {
	const { children, setNodeStore, ...other } = props;
	const [nodes, setNodes] = useState(0);

	useEffect(() => {
		let t = null;
		let pull = () => {
			let intv = props.interval || 15 * 1000;

			// console.log('[pull][node]', intv);
			fetch('./api/node/').then(function (res) {
				return res.json();
			}).then(function (d) {
				// console.log(d);
				d.sort((a,b) => a.tag.localeCompare(b.tag));
				setNodes(d);
				setNodeStore(d);
			});
			t = setTimeout(pull, intv);
		};
		pull();
		return () => {
			clearTimeout(t);
			// console.log('[pull][node]stop');
		};
	}, [props.interval]);

	return (
		<div>
			<DataList header={header} renderRow={renderRow} data={nodes}></DataList>
		</div>
	);
}

export { NodePanel };
