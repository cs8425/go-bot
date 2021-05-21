
import { h, Fragment, Component, render } from 'preact';
import { useState, useEffect, useContext } from 'preact/hooks';

import { makeStyles } from '@material-ui/core/styles';

// import Button from '@material-ui/core/Button';
import Fab from '@material-ui/core/Fab';
import Tooltip from '@material-ui/core/Tooltip';
import AddIcon from '@material-ui/icons/Add';
import ClearIcon from '@material-ui/icons/Clear';

import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';
import { DataList } from './comp.jsx';

const useStyles = makeStyles((theme) => ({
	addBtn: {
		margin: theme.spacing(2),
	},
}));

const header = (
	<TableRow>
		<TableCell key='act'>op</TableCell>
		<TableCell
			key='id'
			style={{ minWidth: 250 }}
		>
			Node ID
		</TableCell>
		<TableCell
			key='addr'
			align='right'
			style={{ minWidth: 150 }}
		>
			Addr
		</TableCell>
		<TableCell
			key='args'
			align='right'
			style={{ minWidth: 150 }}
		>
			Options
		</TableCell>
	</TableRow>
);

const renderRow = (v, idx) => {
	return (
		<TableRow hover role="checkbox" tabIndex={-1} key={idx}>
			<TableCell key='op'>
				<Tooltip title="Stop" aria-label="stop">
					<Fab size="small" color="secondary">
						<ClearIcon />
					</Fab>
				</Tooltip>
			</TableCell>
			<TableCell key='id'>{v.id}</TableCell>
			<TableCell key='addr' align='right'>{v.addr}</TableCell>
			<TableCell key='args' align='right'>{v.args?.join(',')}</TableCell>
		</TableRow>
	);
}

function LocalPanel(props) {
	const classes = useStyles();
	const { children, NodeStore, ...other } = props;
	const [loSrv, setLoSrv] = useState(0);

	const store = useContext(NodeStore);

	useEffect(() => {
		let t = null;
		let pull = () => {
			let intv = props.interval || 15 * 1000;

			console.log('[pull][local]', intv, store);
			fetch('./api/local/').then(function (res) {
				return res.json();
			}).then(function (d) {
				// console.log(d);
				setLoSrv(d);
			});
			t = setTimeout(pull, intv);
		};
		pull();
		return () => {
			clearTimeout(t);
			console.log('[pull][local]stop');
		};
	}, [props.interval, store]);

	return (
		<div>
			<Tooltip title="Add" aria-label="add">
				<Fab color="primary" className={classes.addBtn}>
					<AddIcon />
				</Fab>
			</Tooltip>
			<DataList header={header} renderRow={renderRow} data={loSrv}></DataList>
		</div>
	);
}

export { LocalPanel };
