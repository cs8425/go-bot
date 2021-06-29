// TODO: move some common part to comp.jsx
import { h, Fragment } from 'preact';
import { useState, useEffect, useRef, useContext } from 'preact/hooks';

import { NodeStore } from './store.js';
import { fetchReq, dumpJson } from './api.js';

import { makeStyles } from '@material-ui/core/styles';

import Tooltip from '@material-ui/core/Tooltip';

import Box from '@material-ui/core/Box';
import Fab from '@material-ui/core/Fab';
import Button from '@material-ui/core/Button';
import ButtonGroup from '@material-ui/core/ButtonGroup';

import TextField from '@material-ui/core/TextField';
import MenuItem from '@material-ui/core/MenuItem';

import AddIcon from '@material-ui/icons/Add';
import ClearIcon from '@material-ui/icons/Clear';
import DeleteSweepIcon from '@material-ui/icons/DeleteSweep';
import EditIcon from '@material-ui/icons/Edit';
import FolderOpenIcon from '@material-ui/icons/FolderOpen';
import SaveIcon from '@material-ui/icons/Save';

import Card from '@material-ui/core/Card';
import CardHeader from '@material-ui/core/CardHeader';

import { AlertDialog, PopoverDialog } from './comp.jsx';
import { DragNdrop } from './dragzone.jsx';

const useStyles = makeStyles((theme) => ({
	addBtn: {
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
	cardContent: {
		'overflow-wrap': 'anywhere',
	},
}));

function PanelListKeys(props) {
	const { children, useStyles, masterKeys, onRemove, ...other } = props;
	const classes = useStyles();
	const [popover, setPopover] = useState(null);

	// popover for stop
	const handleClick = (ev, val) => {
		setPopover({
			el: ev.currentTarget,
			val: val,
		});
	};
	const handleStop = () => {
		const val = popover.val;
		const err = onRemove(val);
		if (err) {
			console.log('[key][rm]err', err);
			setDialog({
				title: 'Error',
				msg: err,
			});
		}
	}

	return (
		<div>
			<PopoverDialog
				data={popover}
				setData={setPopover}
				onConfirm={handleStop}
			>
				<p>確定要移除嗎?</p>
			</PopoverDialog>

			{masterKeys.map((v, i) => {
				return (
					<Card className={classes.card} variant="outlined">
						<CardHeader classes={{
							// root: classes.card,
							// action: classes.cardAction,
							avatar: classes.cardAction,
							content: classes.cardContent,
						}}
							avatar={
								<Tooltip title="Option" aria-label="option">
									<Fab size="small" color="secondary" onClick={(e) => handleClick(e, v)}>
										<EditIcon />
									</Fab>
								</Tooltip>
							}
							title={v.tag}
							titleTypographyProps={{ 'variant': 'h5' }}
							subheader={v.note || null}
						/>
					</Card>);
			})}

		</div>
	);
}


function KeyEditPanel(props) {
	const classes = useStyles();
	const { children, ...other } = props;
	const store = useContext(NodeStore);
	const [isAddMode, setAddMode] = useState(false);

	const fileRef = useRef();
	const dummyDlEl = useRef(null);
	const [masterKeys, setMasterKeys] = useState([]);

	// TODO: merge to one State
	const [useNode, setUseNode] = useState('');
	const [masterKey, setMasterKey] = useState(null);

	const [dialogData, setDialog] = useState(null);
	const [popover, setPopover] = useState(null);

	// add req & cancel
	const handleCancel = () => {
		setMasterKey('');
		setAddMode(false);
	}
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

		handleCancel();
	}
	const handleRemove = (v) => {
		console.log('[key][rm]', v);
	}

	const handleLoadBtn = (e) => {
		// console.log('[load]click', e);
		fileRef.current.open();
	}
	const handleFile = (val) => {
		// console.log('[file]', val);
		let reader = new FileReader();
		reader.onload = (e) => {
			const json = JSON.parse(e.target.result);
			console.log(e, json);

			setMasterKeys(json?.keys);
		}
		reader.readAsText(val[0]);
	};
	const handleSave = (e) => {
		dumpJson(dummyDlEl.current, {
			keys: masterKeys,
		}, 'keys.json');
	}

	const handleClearBtn = (e) => {
		setPopover({
			el: e.currentTarget,
		});
	}
	const handleClear = (e) => {
		setMasterKeys([]);
	}

	return (
		<div>
			<AlertDialog data={dialogData} setDialog={setDialog}></AlertDialog>
			{!isAddMode &&
				<DragNdrop ref={fileRef} handleFile={handleFile} onClick={false}>
					<Tooltip title="Load" aria-label="load">
						<Fab color="primary" className={classes.addBtn} onClick={handleLoadBtn}>
							<FolderOpenIcon />
						</Fab>
					</Tooltip>
					<Tooltip title="Save" aria-label="save">
						<Fab color="primary" className={classes.addBtn} onClick={handleSave}>
							<SaveIcon />
						</Fab>
					</Tooltip>

					<Tooltip title="Add" aria-label="add">
						<Fab color="primary" className={classes.addBtn} onClick={() => setAddMode(true)}>
							<AddIcon />
						</Fab>
					</Tooltip>
					<Tooltip title="Clear all" aria-label="clear all">
						<Fab color="secondary" className={classes.addBtn} onClick={handleClearBtn}>
							<DeleteSweepIcon />
						</Fab>
					</Tooltip>
					<PanelListKeys
						useStyles={useStyles}
						onRemove={handleRemove}
						masterKeys={masterKeys}
						setMasterKeys={setMasterKeys}
					></PanelListKeys>

					<PopoverDialog
						data={popover}
						setData={setPopover}
						onConfirm={handleClear}
					>
						<p>確定要全部移除嗎?</p>
					</PopoverDialog>
				</DragNdrop>
			}
			{isAddMode &&
				<Box className={classes.center}>
					<div style="margin: 1rem;">
						<TextField
							required
							select
							label="Node"
							value={useNode}
							onChange={(e) => setUseNode(e.target.value)}
							helperText="Please select a using node"
						>
							<MenuItem value={''}>---</MenuItem>
							{store.map((option) => (
								<MenuItem key={option.tag} value={option.tag}>
									{option.tag}
								</MenuItem>
							))}
						</TextField>
					</div>
					<div style="margin: 1rem;">
						<TextField
							required
							// fullWidth
							label="Node"
							value={useNode}
							onChange={(e) => setUseNode(e.target.value)}
							InputLabelProps={{ shrink: !!useNode }}
							helperText="Or input a node"
						/>
					</div>
					<hr />

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
							<Button className={classes.noUppercase} onClick={handleCancel}>Cancel</Button>
							<Button className={classes.noUppercase} onClick={handleAdd} color="primary" >Add</Button>
						</ButtonGroup>
					</div>
				</Box>
			}

			{/* dummy link for download file */}
			<a style="display: none;" ref={dummyDlEl}></a>
		</div>
	);
}

export { KeyEditPanel };
