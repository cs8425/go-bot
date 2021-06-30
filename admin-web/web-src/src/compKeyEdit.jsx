// TODO: move some common part to comp.jsx
import { h, Fragment } from 'preact';
import { useState, useRef, useContext } from 'preact/hooks';

import { NodeStore } from './store.js';
import { dumpJson } from './api.js';

import { makeStyles } from '@material-ui/core/styles';

import Box from '@material-ui/core/Box';
import Fab from '@material-ui/core/Fab';
import Button from '@material-ui/core/Button';
import ButtonGroup from '@material-ui/core/ButtonGroup';

import TextField from '@material-ui/core/TextField';
import MenuItem from '@material-ui/core/MenuItem';
import Tooltip from '@material-ui/core/Tooltip';
import Typography from '@material-ui/core/Typography';

import AddIcon from '@material-ui/icons/Add';
import AddCircleIcon from '@material-ui/icons/AddCircle';
import CancelIcon from '@material-ui/icons/Cancel';
import DeleteIcon from '@material-ui/icons/Delete';
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
	text: {
		color: theme.palette.text.primary,
	},
}));

function PanelListKeys(props) {
	const { children, masterKeys, onRemove, onEdit, ...other } = props;
	const classes = useStyles();

	const handleClick = (ev, val, i) => {
		if (typeof onEdit === 'function') return onEdit(ev, val, i);
	};

	return (
		<>
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
								<Tooltip title="Edit" aria-label="edit">
									<Fab size="small" color="secondary" onClick={(e) => handleClick(e, v, i)}>
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
		</>
	);
}

function KeyEdit(props) {
	const { children, editData, setEditData, isNew, onCancel, onAdd, onSave, onRemove, ...other } = props;
	const classes = useStyles();
	const store = useContext(NodeStore);
	const [dialogData, setDialog] = useState(null);

	const handleFn = (fn) => (e) => {
		if (typeof fn === 'function') fn(e, editData);
	}
	const handleRemove = (e) => {
		setDialog({
			msg: (
				<Typography variant="h6" className={classes.text} >確定要刪除?</Typography>
			),
		});
	}

	return (
		<Box className={classes.center}>
			<AlertDialog
				data={dialogData}
				setDialog={setDialog}
				footer={
					<>
						<Button className={classes.noUppercase} onClick={() => setDialog(null)} color="primary">Cancel</Button>
						<Button className={classes.noUppercase} onClick={handleFn(onRemove)} color="secondary">Remove</Button>
					</>
				}
			/>
			<div style="margin: 1rem;">
				<TextField
					required
					select
					label="Node"
					value={editData.node || ''}
					onChange={(e) => setEditData({ ...editData, node: e.target.value })}
					helperText="Please select a node"
				>
					<MenuItem value={''}>---</MenuItem>
					{store.map((option) => {
						const tag = option.tag.split('/')[0];
						return (<MenuItem key={tag} value={tag}>{tag}</MenuItem>);
					})}
				</TextField>
			</div>
			<div style="margin: 1rem;">
				<TextField
					required
					// fullWidth
					label="Node"
					value={editData.node}
					onChange={(e) => setEditData({ ...editData, node: e.target.value })}
					InputLabelProps={{ shrink: !!editData.node }}
					helperText="Or input a node"
				/>
			</div>

			<div style="margin: 1rem;">
				<TextField
					multiline
					label="Note"
					value={editData.note}
					onChange={(e) => setEditData({ ...editData, note: e.target.value })}
					helperText="註解"
				/>
			</div>

			{/* <hr /> */}

			{/* TODO: or from file */}
			<div style="margin: 1rem;">
				<TextField
					required
					fullWidth
					label="Master Key (base64)"
					value={editData.key}
					onChange={(e) => setEditData({ ...editData, key: e.target.value })}
				/>
			</div>
			<div style="margin: 2rem;">
				{isNew &&
					<ButtonGroup disableElevation variant="contained" fullWidth="true">
						<Button className={classes.noUppercase} onClick={handleFn(onCancel)}><CancelIcon />Cancel</Button>
						<Button className={classes.noUppercase} onClick={handleFn(onAdd)} color="primary" ><AddCircleIcon />Add</Button>
					</ButtonGroup>
				}
				{!isNew &&
					<ButtonGroup disableElevation variant="contained" fullWidth="true">
						<Button className={classes.noUppercase} onClick={handleFn(onCancel)}><CancelIcon />Cancel</Button>
						<Button className={classes.noUppercase} onClick={handleFn(onSave)} color="primary" ><SaveIcon />Save</Button>
						<Button className={classes.noUppercase} onClick={handleRemove} color="secondary" ><DeleteIcon />Remove</Button>
					</ButtonGroup>
				}
			</div>
		</Box>
	);
}

function KeyEditPanel(props) {
	const classes = useStyles();
	const { children, ...other } = props;
	const [editMode, setEditMode] = useState(0); // 0: list, 1: new, 2: edit

	const fileRef = useRef();
	const dummyDlEl = useRef(null);
	const [dialogData, setDialog] = useState(null);
	const [popover, setPopover] = useState(null);

	const [masterKeys, setMasterKeys] = useState([]);
	const [editData, setEditData] = useState({});

	// add req & cancel
	const handleCancel = () => {
		setEditData({});
		setEditMode(0);
	}
	const verify = (v) => {
		if (!v?.node) {
			setDialog({
				title: '請填入節點!!',
			});
			return;
		}
		if (!v?.key) {
			setDialog({
				title: '請填入key!!',
			});
			return;
		}
		return true;
	}
	const handleAdd = (e, v) => {
		if (!verify(v)) return;
		let param = {
			tag: v?.node?.split('/')[0],
			key: v?.key,
			note: v?.note,
		};
		// console.log('[key][add]', editData, v, param);

		let list = [...masterKeys, param];
		setMasterKeys(list);
		handleCancel();
	}
	const handleRemove = (e, v) => {
		let list = masterKeys.filter((s, i) => i !== v.idx); // by index
		setMasterKeys(list);
		// console.log('[key][rm]', v, list, masterKeys);
		handleCancel();
	}
	const handleEditSave = (e, v) => {
		// console.log('[key][save]', v);
		if (!verify(v)) return;
		const tag = v?.node?.split('/')[0];
		let item = masterKeys[v.idx]; // by index
		if (!item) return;
		Object.assign(item, {
			tag,
			key: v.key,
			note: v.note,
		});
		setMasterKeys([...masterKeys]);

		handleCancel();
	}

	const handleEditMode = (e, v, i) => {
		// console.log('[key][edit]', e, v, i);
		setEditData({
			node: v.tag,
			key: v.key,
			note: v.note,
			idx: i,
		});
		setEditMode(2);
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
			// console.log(e, json);

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
			{(editMode === 0) &&
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
						<Fab color="primary" className={classes.addBtn} onClick={() => setEditMode(1)}>
							<AddIcon />
						</Fab>
					</Tooltip>
					<Tooltip title="Clear all" aria-label="clear all">
						<Fab color="secondary" className={classes.addBtn} onClick={handleClearBtn}>
							<DeleteSweepIcon />
						</Fab>
					</Tooltip>
					<PanelListKeys
						onRemove={handleRemove}
						onEdit={handleEditMode}
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
			{(editMode !== 0) &&
				<KeyEdit
					isNew={editMode == 1}
					editData={editData}
					setEditData={setEditData}
					onCancel={handleCancel}
					onAdd={handleAdd}
					onSave={handleEditSave}
					onRemove={handleRemove}
				/>
			}

			{/* dummy link for download file */}
			<a style="display: none;" ref={dummyDlEl}></a>
		</div>
	);
}

export { KeyEditPanel };
