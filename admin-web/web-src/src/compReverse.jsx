// TODO: move some common part to comp.jsx
import { h, Fragment, Component, render } from 'preact';
import { useState, useEffect, useContext } from 'preact/hooks';

import { makeStyles } from '@material-ui/core/styles';

import Tooltip from '@material-ui/core/Tooltip';
import Popover from '@material-ui/core/Popover';

import Box from '@material-ui/core/Box';
import Fab from '@material-ui/core/Fab';
import Button from '@material-ui/core/Button';
import ButtonGroup from '@material-ui/core/ButtonGroup';
import Switch from '@material-ui/core/Switch';

import TextField from '@material-ui/core/TextField';
import MenuItem from '@material-ui/core/MenuItem';

import AddIcon from '@material-ui/icons/Add';
import ClearIcon from '@material-ui/icons/Clear';

import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';
import { DataList, AlertDialog } from './comp.jsx';

const useStyles = makeStyles((theme) => ({
	addBtn: {
		margin: theme.spacing(2),
	},
	popover: {
		margin: theme.spacing(2),
	},
	center: {
		textAlign: 'center',
	},
	formControl: {
		margin: theme.spacing(1),
		minWidth: 120,
	},
	bindType: {
		minWidth: 180,
	},
	noUppercase: {
		textTransform: 'unset',
	},
}));

const header = (
	<TableRow>
		<TableCell key='act'>op</TableCell>
		<TableCell key='ks'>Pause</TableCell>
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
			key='target'
			align='right'
			style={{ minWidth: 150 }}
		>
			Target
		</TableCell>
	</TableRow>
);

function PanelListMode(props) {
	const classes = useStyles();
	const { children, handleAddBtn, ...other } = props;
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
		fetch(`./api/rev/?op=stop&cid=${val.cid}`, {
			method: 'POST',
		}).then((res) => {
			return res.json();
		}).then((d) => {
			console.log('[rev][stop]', d);
			setAnchorEl(null);
			setLoSrv(d);
		}).finally(() => {
			setAnchorEl(null);
		});
	}
	const handleKS = (e, val) => {
		console.log('[KS]', e, val);
		const ks = (val.pause) ? '0':'1';
		fetch(`./api/rev/?op=ks&cid=${val.cid}&val=${ks}`, {
			method: 'POST',
		}).then((res) => {
			return res.json();
		}).then((d) => {
			console.log('[rev][ks]', d);
			setLoSrv(d);
		});
	}

	useEffect(() => {
		let t = null;
		let pull = () => {
			let intv = props.interval || 15 * 1000;

			// console.log('[pull][rev]', intv);
			fetch('./api/rev/').then(function (res) {
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
			console.log('[pull][rev]stop');
		};
	}, [props.interval]);

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
				<TableCell key='ks'>
					<Switch color="primary" checked={v.pause} onChange={(e) => handleKS(e, v)} name="pause" />
				</TableCell>
				<TableCell key='id'>{v.id}</TableCell>
				<TableCell key='addr' align='right'>{v.addr}</TableCell>
				<TableCell key='target' align='right'>{v.target}</TableCell>
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
						<Button className={classes.noUppercase} onClick={handleClose}>Cancel</Button>
						<Button className={classes.noUppercase} onClick={handleStop} color="secondary" >Stop</Button>
					</ButtonGroup>
				</Box>
			</Popover>

			<Tooltip title="Add" aria-label="add">
				<Fab color="primary" className={classes.addBtn} onClick={handleAddBtn}>
					<AddIcon />
				</Fab>
			</Tooltip>
			<DataList header={header} renderRow={renderRow} data={loSrv}></DataList>
		</div>
	);
}


function ReversePanel(props) {
	const classes = useStyles();
	const { children, NodeStore, ...other } = props;
	const store = useContext(NodeStore);
	const [isAddMode, setAddMode] = useState(false);

	const [useNode, setUseNode] = useState(null);
	const [bindType, setBindType] = useState('any');
	const [bindPort, setBindPort] = useState(19000);
	const [bindAddr, setBindAddr] = useState('127.0.0.1');
	const [targetAddr, setTargetAddr] = useState('127.0.0.1:443');

	const [dialogData, setDialog] = useState(null);

	// add req
	const handleAdd = (e) => {
		if (!useNode) {
			// alert
			setDialog({
				title: '請選擇節點!!',
			});
			return;
		}
		let param = {
			uuid: useNode,
			remote: '',
			target: targetAddr,
		};
		switch (bindType) {
			case 'local':
				param.remote = `127.0.0.1:${bindPort}`;
				break;
			case 'any':
				param.remote = `:${bindPort}`;
				break;
			case 'custom':
				param.remote = `${bindAddr}:${bindPort}`;
				break;
		}

		console.log('[rev][add]', param);
		fetch('./api/rev/?op=bind', {
			body: JSON.stringify(param),
			method: 'POST',
		}).then(function (res) {
			return res.json();
		}).then(function (d) {
			console.log('[rev][add]ret', d);
			setAddMode(false);
		});
	}

	return (
		<div>
			{ !isAddMode &&
				<PanelListMode handleAddBtn={() => setAddMode(true)}></PanelListMode>
			}
			{ isAddMode &&
				<Box className={classes.center}>
					<AlertDialog data={dialogData} setDialog={setDialog}></AlertDialog>
					<div style="margin: 1rem;">
						<TextField
							required
							select
							label="Node"
							value={useNode}
							onChange={(e) => setUseNode(e.target.value)}
							helperText="Please select a using node"
						>
							<MenuItem value={null}>---</MenuItem>
							{store.map((option) => (
								<MenuItem key={option.tag} value={option.tag}>
									{option.tag}
								</MenuItem>
							))}
						</TextField>
					</div>
					<div style="margin: 1rem;">
						<TextField
							className={classes.bindType}
							required
							select
							label="type"
							value={bindType}
							onChange={(e) => setBindType(e.target.value)}
						>
							<MenuItem key='local' value='local'>Local</MenuItem>
							<MenuItem key='any' value='any'>Any</MenuItem>
							<MenuItem key='custom' value='custom'>Custom</MenuItem>
						</TextField>
						{(bindType == 'custom') &&
							<TextField
								required
								label="addr"
								value={bindAddr}
								onChange={(e) => setBindAddr(e.target.value)}
							/>
						}
						<TextField
							required
							type="number"
							label="port"
							value={bindPort}
							onChange={(e) => setBindPort(parseInt(e.target.value))}
							helperText="0 for auto port"
						/>
					</div>
					<div style="margin: 1rem;">
						<TextField
							required
							label="target addr"
							value={targetAddr}
							onChange={(e) => setTargetAddr(e.target.value)}
						/>
					</div>
					<div style="margin: 2rem;">
						<ButtonGroup disableElevation variant="contained" fullWidth="true">
							<Button className={classes.noUppercase} onClick={() => setAddMode(false)}>Cancel</Button>
							<Button className={classes.noUppercase} onClick={handleAdd} color="primary" >Add</Button>
						</ButtonGroup>
					</div>
				</Box>
			}
		</div>
	);
}

export { ReversePanel };
