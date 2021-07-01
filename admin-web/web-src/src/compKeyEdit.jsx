// TODO: move some common part to comp.jsx
import { h, Fragment } from 'preact';
import { useState, useRef } from 'preact/hooks';

import { dumpJson, cryptoApi } from './api.js';

import { makeStyles } from '@material-ui/core/styles';

import Fab from '@material-ui/core/Fab';
import Tooltip from '@material-ui/core/Tooltip';

import AddIcon from '@material-ui/icons/Add';
import DeleteSweepIcon from '@material-ui/icons/DeleteSweep';
import EditIcon from '@material-ui/icons/Edit';
import FolderOpenIcon from '@material-ui/icons/FolderOpen';
import SaveIcon from '@material-ui/icons/Save';

import Card from '@material-ui/core/Card';
import CardHeader from '@material-ui/core/CardHeader';

import { AlertDialog, PopoverDialog } from './comp.jsx';
import { DragNdrop } from './dragzone.jsx';
import { KeyEdit, PwdDialog } from './compUI.jsx';

const useStyles = makeStyles((theme) => ({
	addBtn: {
		margin: theme.spacing(2),
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

function KeyEditPanel(props) {
	const classes = useStyles();
	const { children, ...other } = props;
	const [editMode, setEditMode] = useState(0); // 0: list, 1: new, 2: edit

	const fileRef = useRef();
	const dummyDlEl = useRef(null);
	const [dialogData, setDialog] = useState(null);
	const [popover, setPopover] = useState(null);
	const [pwdDialog, setPwdDialog] = useState(null);

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
		if (!val.length) return;
		let reader = new FileReader();
		reader.onload = (e) => {
			const json = JSON.parse(e.target.result);
			// console.log(e, json);

			// Dialog for password
			if (cryptoApi.isEncrypt(json)) {
				console.log('[file]encrypted', json);
				setPwdDialog({
					title: '檔案已加密',
					cb: (ev, val) => handleFileDec(json, val),
				});
				return;
			}

			setMasterKeys(json?.keys);
		}
		reader.readAsText(val[0]);
	};
	const handleFileDec = async (json, pwd) => {
		console.log('[pwd]', pwd, json);
		try { // decrypt data
			const obj = await cryptoApi.decrypt(json, pwd);
			console.log('[dec]', obj);
			setMasterKeys(obj?.keys);
		} catch (err) {
			setDialog({
				title: 'Error',
				msg: err.toString(),
			});
		}
	}
	const handleSave = async (e) => {
		const dump = {
			keys: masterKeys,
		};
		// dumpJson(dummyDlEl.current, dump, 'keys.json');

		const handleSaveEnc = async (dump, pwd) => {
			if (pwd === '') dumpJson(dummyDlEl.current, dump, 'keys.json');
			const obj = await cryptoApi.encrypt(dump, pwd);
			console.log('[enc]', obj);
			dumpJson(dummyDlEl.current, obj, 'keys.json');
		}

		//  Dialog for password
		setPwdDialog({
			title: '請輸入加密密碼',
			setPwd: true,
			cb: (ev, val) => handleSaveEnc(dump, val),
		});
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
			<AlertDialog data={dialogData} setDialog={setDialog} />
			<PwdDialog data={pwdDialog} setDialog={setPwdDialog} setPwd={pwdDialog?.setPwd} />
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
