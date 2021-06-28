// TODO: move some common part to comp.jsx
import { h, Fragment } from 'preact';
import { useState, useEffect, useContext } from 'preact/hooks';

import { NodeStore, KeyStore } from './store.js';
import { fetchReq } from './api.js';

import { makeStyles } from '@material-ui/core/styles';

import Tooltip from '@material-ui/core/Tooltip';
import Popover from '@material-ui/core/Popover';

import Box from '@material-ui/core/Box';
import Fab from '@material-ui/core/Fab';
import Button from '@material-ui/core/Button';
import ButtonGroup from '@material-ui/core/ButtonGroup';

import TextField from '@material-ui/core/TextField';
import MenuItem from '@material-ui/core/MenuItem';

import AddIcon from '@material-ui/icons/Add';
import ClearIcon from '@material-ui/icons/Clear';

import Card from '@material-ui/core/Card';
import CardHeader from '@material-ui/core/CardHeader';

import { AlertDialog } from './comp.jsx';

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
	noUppercase: {
		textTransform: 'unset',
	},
	card: {
		margin: theme.spacing(2),
	},
	cardAction: {
		margin: '0 1rem 0 0',
	},
}));

function PanelListKeys(props) {
	const { children, useStyles, handleAddBtn, stopParamFn, pullFn, dataStore, ...other } = props;
	const classes = useStyles();
	const [anchorEl, setAnchorEl] = useState(null);
	// const srvStore = useContext(dataStore);
	const [masterKeys, setMasterKeys] = useState([]);

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
		console.log('[rm]', anchorEl.val);
		const val = anchorEl.val;
		const param = stopParamFn(val);
		fetchReq(param.url, param.param, (d) => {
			console.log('[rm][stop]', d);
			setAnchorEl(null);
			// srvStore.set(d);
			setMasterKeys(d);
		}, (err) => {
			console.log('[rm][stop]err', err);
			setDialog({
				title: 'Error',
				msg: err,
			});
		});
	}

	useEffect(() => {
		let t = null;
		let intv = props.interval || 15 * 1000;
		let pull = () => {
			// console.log('[pull][key]', intv);
			pullFn()?.then((d) => {
				console.log('[pull][key]', d);
				// srvStore.set(d);
				setMasterKeys(d);
			});
			t = setTimeout(pull, intv);
		};
		pull();
		return () => {
			clearTimeout(t);
			// console.log('[pull][key]stop');
		};
	}, [props.interval]);

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
					<p>確定要移除嗎?</p>
					<ButtonGroup disableElevation variant="contained">
						<Button className={classes.noUppercase} onClick={handleClose}>Cancel</Button>
						<Button className={classes.noUppercase} onClick={handleStop} color="secondary" >Remove</Button>
					</ButtonGroup>
				</Box>
			</Popover>

			<Tooltip title="Add" aria-label="add">
				<Fab color="primary" className={classes.addBtn} onClick={handleAddBtn}>
					<AddIcon />
				</Fab>
			</Tooltip>

			{masterKeys.map((v, i) => {
				return (
					<Card className={classes.card} variant="outlined">
						<CardHeader classes={{
							// root: classes.card,
							// action: classes.cardAction,
							avatar: classes.cardAction,
						}}
							avatar={
								<Tooltip title="Remove" aria-label="remove">
									<Fab size="small" color="secondary" onClick={(e) => handleClick(e, v)}>
										<ClearIcon />
									</Fab>
								</Tooltip>
							}
							// action={
							// 	<Tooltip title="Remove" aria-label="remove">
							// 		<Fab size="small" color="secondary" onClick={(e) => handleClick(e, v)}>
							// 			<ClearIcon />
							// 		</Fab>
							// 	</Tooltip>
							// }
							title={v}
							titleTypographyProps={{ 'variant': 'h5' }}
						/>
					</Card>);
			})}

		</div>
	);
}


function KeyPanel(props) {
	const classes = useStyles();
	const { children, ...other } = props;
	const store = useContext(NodeStore);
	const [isAddMode, setAddMode] = useState(false);

	// TODO: merge to one State
	const [useNode, setUseNode] = useState(null);
	const [masterKey, setMasterKey] = useState(null);

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
			uuid: useNode?.split('/')[0],
			key: masterKey,
		};
		console.log('[key][add]', param);
		fetchReq('./api/key/?op=set', {
			body: JSON.stringify(param),
			method: 'POST',
		}, (d) => {
			console.log('[key][add]ret', d);
			setAddMode(false);
		}, (err) => {
			console.log('[key][add]err', err);
			setDialog({
				title: 'Error',
				msg: err,
			});
		});
	}

	const stopParamFn = (val) => {
		return {
			url: `./api/key/?op=rm`,
			param: {
				body: JSON.stringify({ uuid: val }),
				method: 'POST',
			},
		};
	}

	const pullFn = () => {
		return new Promise((resolve, reject) => {
			fetchReq('./api/key/').then((d) => {
				if (d.sort) d.sort();
				resolve(d);
			})
		});
	};

	return (
		<div>
			<AlertDialog data={dialogData} setDialog={setDialog}></AlertDialog>
			{!isAddMode &&
				<PanelListKeys
					handleAddBtn={() => setAddMode(true)}
					useStyles={useStyles}
					stopParamFn={stopParamFn}
					pullFn={pullFn}
					dataStore={KeyStore}
				></PanelListKeys>
			}
			{isAddMode &&
				<Box className={classes.center}>
					<div style="margin: 1rem;">
						{/* TODO: or from text input */}
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

					{/* TODO: or from file */}
					<div style="margin: 1rem;">
						<TextField
							required
							fullWidth
							label="Master Key (base64)"
							value={masterKey}
							onChange={(e) => setMasterKey(e.target.value)}
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

export { KeyPanel };
