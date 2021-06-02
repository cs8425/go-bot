
// TODO: move some common part to comp.jsx
import { h, Fragment } from 'preact';
import { useState, useContext } from 'preact/hooks';

import { NodeStore, LocalStore } from './store.js';
import { fetchReq } from './api.js';

import { makeStyles } from '@material-ui/core/styles';

import Tooltip from '@material-ui/core/Tooltip';

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

import ClearIcon from '@material-ui/icons/Clear';

import TableCell from '@material-ui/core/TableCell';
import TableRow from '@material-ui/core/TableRow';
import { AlertDialog, PanelListMode } from './comp.jsx';

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

const renderRow = ({ v, idx, onClick, onKillSwitch }) => {
	return (
		<TableRow hover role="checkbox" tabIndex={-1} key={idx}>
			<TableCell key='op'>
				<Tooltip title="Stop" aria-label="stop">
					<Fab size="small" color="secondary" onClick={onClick}>
						<ClearIcon />
					</Fab>
				</Tooltip>
			</TableCell>
			<TableCell key='ks'>
				<Switch color="primary" checked={v.pause} onChange={onKillSwitch} name="pause" />
			</TableCell>
			<TableCell key='id'>{v.id}</TableCell>
			<TableCell key='addr' align='right'>{v.addr}</TableCell>
			<TableCell key='args' align='right'>{v.args?.join(',')}</TableCell>
		</TableRow>
	);
}

function LocalPanel(props) {
	const classes = useStyles();
	const { children, ...other } = props;
	const store = useContext(NodeStore);
	const [isAddMode, setAddMode] = useState(false);

	// TODO: merge to one State
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

		console.log('[local][add]', param);
		fetchReq('./api/local/?op=bind', {
			body: JSON.stringify(param),
			method: 'POST',
		}, (d) => {
			console.log('[local][add]ret', d);
			setAddMode(false);
		}, (err) => {
			console.log('[local][add]err', err);
			setDialog({
				title: 'Error',
				msg: err,
			});
		});
	}

	const stopParamFn = (val) => {
		return {
			url: `./api/local/?op=stop&addr=${val.addr}`,
			param: {
				method: 'POST',
			},
		};
	}
	const ksParamFn = (val) => {
		const ks = (val.pause) ? '0' : '1';
		return {
			url: `./api/local/?op=ks&addr=${val.addr}&val=${ks}`,
			param: {
				method: 'POST',
			},
		};
	}

	return (
		<div>
			{ !isAddMode &&
				<PanelListMode
					handleAddBtn={() => setAddMode(true)}
					useStyles={useStyles}
					stopParamFn={stopParamFn}
					ksParamFn={ksParamFn}
					pullFn={() => fetchReq('./api/local/')}
					header={header}
					renderRowFn={renderRow}
					dataStore={LocalStore}
				></PanelListMode>
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
