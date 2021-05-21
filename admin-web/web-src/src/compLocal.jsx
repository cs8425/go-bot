
import { h, Fragment, Component, render } from 'preact';
import { useState, useEffect, useContext } from 'preact/hooks';

import { makeStyles } from '@material-ui/core/styles';

import Tooltip from '@material-ui/core/Tooltip';
import Popover from '@material-ui/core/Popover';

import Box from '@material-ui/core/Box';
import Fab from '@material-ui/core/Fab';
import Button from '@material-ui/core/Button';
import ButtonGroup from '@material-ui/core/ButtonGroup';

import AddIcon from '@material-ui/icons/Add';
import ClearIcon from '@material-ui/icons/Clear';

import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';
import { DataList } from './comp.jsx';

const useStyles = makeStyles((theme) => ({
	addBtn: {
		margin: theme.spacing(2),
	},
	popover: {
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



function LocalPanel(props) {
	const classes = useStyles();
	const { children, NodeStore, ...other } = props;
	const store = useContext(NodeStore);
	const [loSrv, setLoSrv] = useState(0);
	const [anchorEl, setAnchorEl] = useState(null);

	// popover for stop
	const handleClick = (ev, val) => {
		console.log('[anchorEl]', ev, val);
		setAnchorEl({
			el: ev.currentTarget,
			val: val,
		});
	};
	const handleClose = () => {
		setAnchorEl(null);
	};
	const handleStop = () => {
		console.log('[stop]', anchorEl.val);
		const val = anchorEl.val;
		fetch(`./api/local/?op=stop&addr=${val.addr}`,{
			method: 'POST',
		})
		.then((res) => {
			return res.json();
		})
		.then((d) => {
			console.log('[local][stop]', d);
			setLoSrv(d);
		}).finally(() => {
			setAnchorEl(null);
		});
	}

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

	const renderRow = (v, idx) => {
		return (
			<TableRow hover role="checkbox" tabIndex={-1} key={idx}>
				<TableCell key='op'>
					<Tooltip title="Stop" aria-label="stop">
						<Fab size="small" color="secondary" onClick={(e) => handleClick(e, v)}>
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

	return (
		<div>
			<Popover
				open={anchorEl !== null}
				onClose={handleClose}
				anchorEl={anchorEl?.el}
				anchorOrigin={{
					vertical: 'top',
					horizontal: 'left',
				}}
				transformOrigin={{
					vertical: 'top',
					horizontal: 'left',
				}}
			>
				<Box className={classes.popover}>
					<p>確定要停止嗎?</p>
					<ButtonGroup disableElevation variant="contained">
						<Button onClick={handleClose}>Cancel</Button>
						<Button onClick={handleStop} color="secondary" >Stop</Button>
					</ButtonGroup>
				</Box>
			</Popover>

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
