// TODO: move some common part to comp.jsx
import { h, Fragment } from 'preact';
import { useState, useContext } from 'preact/hooks';

import { NodeStore, RevStore } from './store.js';
import { fetchReq } from './api.js';

import { makeStyles } from '@material-ui/core/styles';

import Tooltip from '@material-ui/core/Tooltip';

import Box from '@material-ui/core/Box';
import Fab from '@material-ui/core/Fab';
import Button from '@material-ui/core/Button';
import ButtonGroup from '@material-ui/core/ButtonGroup';
import Switch from '@material-ui/core/Switch';

import TextField from '@material-ui/core/TextField';
import MenuItem from '@material-ui/core/MenuItem';

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
			key='target'
			align='right'
			style={{ minWidth: 150 }}
		>
			Target
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
			<TableCell key='target' align='right'>{v.target}</TableCell>
		</TableRow>
	);
}

function ReversePanel(props) {
	const classes = useStyles();
	const { children, ...other } = props;
	const store = useContext(NodeStore);
	const [isAddMode, setAddMode] = useState(false);

	// TODO: merge to one State
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
		fetchReq('./api/rev/?op=bind', {
			body: JSON.stringify(param),
			method: 'POST',
		}, (d) => {
			console.log('[rev][add]ret', d);
			setAddMode(false);
		}, (err) => {
			console.log('[rev][add]err', err);
			setDialog({
				title: 'Error',
				msg: err,
			});
		});
	}

	const stopParamFn = (val) => {
		return {
			url: `./api/rev/?op=stop&cid=${val.cid}`,
			param: {
				method: 'POST',
			},
		};
	}
	const ksParamFn = (val) => {
		const ks = (val.pause) ? '0' : '1';
		return {
			url: `./api/rev/?op=ks&cid=${val.cid}&val=${ks}`,
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
					pullFn={() => fetchReq('./api/rev/')}
					header={header}
					renderRowFn={renderRow}
					dataStore={RevStore}
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
