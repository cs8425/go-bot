
// TODO: move some common part to comp.jsx
import { h, Fragment, Component, render } from 'preact';
import { useState, useEffect, useContext } from 'preact/hooks';

import { NodeStore, LocalStore } from './store.js';

import { makeStyles } from '@material-ui/core/styles';

import Tooltip from '@material-ui/core/Tooltip';
import Popover from '@material-ui/core/Popover';

import Box from '@material-ui/core/Box';
import Fab from '@material-ui/core/Fab';
import Button from '@material-ui/core/Button';
import ButtonGroup from '@material-ui/core/ButtonGroup';

import Radio from '@material-ui/core/Radio';
import RadioGroup from '@material-ui/core/RadioGroup';
import TextField from '@material-ui/core/TextField';
import MenuItem from '@material-ui/core/MenuItem';
import FormControlLabel from '@material-ui/core/FormControlLabel';
import FormControl from '@material-ui/core/FormControl';
import FormLabel from '@material-ui/core/FormLabel';
import Switch from '@material-ui/core/Switch';

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
			key='args'
			align='right'
			style={{ minWidth: 150 }}
		>
			Options
		</TableCell>
	</TableRow>
);

function LocalPanelListMode(props) {
	const classes = useStyles();
	const { children, handleAddBtn, ...other } = props;
	const [loSrv, setLoSrv] = useState(0);
	const [anchorEl, setAnchorEl] = useState(null);
	const srvStore = useContext(LocalStore);

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
		fetch(`./api/local/?op=stop&addr=${val.addr}`, {
			method: 'POST',
		}).then((res) => {
			return res.json();
		}).then((d) => {
			console.log('[local][stop]', d);
			setAnchorEl(null);
			srvStore.set(d);
		}).finally(() => {
			setAnchorEl(null);
		});
	}
	const handleKS = (e, val) => {
		console.log('[KS]', e, val);
		const ks = (val.pause) ? '0':'1';
		fetch(`./api/local/?op=ks&addr=${val.addr}&val=${ks}`, {
			method: 'POST',
		}).then((res) => {
			return res.json();
		}).then((d) => {
			console.log('[local][ks]', d);
			srvStore.set(d);
		});
	}

	useEffect(() => {
		let t = null;
		let pull = () => {
			let intv = props.interval || 15 * 1000;

			// console.log('[pull][local]', intv);
			fetch('./api/local/').then(function (res) {
				return res.json();
			}).then(function (d) {
				// console.log(d);
				srvStore.set(d);
			});
			t = setTimeout(pull, intv);
		};
		pull();
		return () => {
			clearTimeout(t);
			// console.log('[pull][local]stop');
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
			<DataList header={header} renderRow={renderRow} data={srvStore.val}></DataList>
		</div>
	);
}


function LocalPanel(props) {
	const classes = useStyles();
	const { children, ...other } = props;
	const store = useContext(NodeStore);
	const [isAddMode, setAddMode] = useState(false);

	const [srvType, setSrvType] = useState('socks');
	const [useNode, setUseNode] = useState(null);
	const [bindType, setBindType] = useState('local');
	const [bindPort, setBindPort] = useState(1080);
	const [bindAddr, setBindAddr] = useState('127.0.0.1');
	const [targetAddr, setTargetAddr] = useState('192.168.1.215:3389');

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
			bind_addr: '',
			argv: [srvType],
		};
		switch (bindType) {
			case 'local':
				param.bind_addr = `127.0.0.1:${bindPort}`;
				break;
			case 'any':
				param.bind_addr = `:${bindPort}`;
				break;
			case 'custom':
				param.bind_addr = `${bindAddr}:${bindPort}`;
				break;
		}
		if (srvType == 'raw') {
			param.argv.push(targetAddr);
		}

		console.log('[loacl][add]', param);
		fetch('./api/local/?op=bind', {
			body: JSON.stringify(param),
			method: 'POST',
		}).then(function (res) {
			return res.json();
		}).then(function (d) {
			console.log('[loacl][add]ret', d);
			setAddMode(false);
		});
	}

	return (
		<div>
			{ !isAddMode &&
				<LocalPanelListMode handleAddBtn={() => setAddMode(true)}></LocalPanelListMode>
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
						<FormControl component="fieldset">
							<FormLabel component="legend">類型</FormLabel>
							<RadioGroup row aria-label="position" name="position" defaultValue={srvType} value={srvType} onChange={(e, v) => setSrvType(v)}>
								<FormControlLabel value="socks" control={<Radio color="primary" />} label="socks5" />
								<FormControlLabel value="http" control={<Radio color="primary" />} label="http" />
								<FormControlLabel value="raw" control={<Radio color="primary" />} label="raw" />
							</RadioGroup>
						</FormControl>
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
					{(srvType == 'raw') &&
						<div style="margin: 1rem;">
							<TextField
								required
								label="target addr"
								value={targetAddr}
								onChange={(e) => setTargetAddr(e.target.value)}
							/>
						</div>
					}
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

export { LocalPanel };
