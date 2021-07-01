// TODO: move some common part to comp.jsx
import { h, Fragment } from 'preact';
import { useState, useEffect, useRef } from 'preact/hooks';

import { fetchReq, cryptoApi } from './api.js';

import { makeStyles } from '@material-ui/core/styles';

import Tooltip from '@material-ui/core/Tooltip';

import Fab from '@material-ui/core/Fab';

import AddIcon from '@material-ui/icons/Add';
import ClearIcon from '@material-ui/icons/Clear';
import DeleteSweepIcon from '@material-ui/icons/DeleteSweep';
import FolderOpenIcon from '@material-ui/icons/FolderOpen';

import Card from '@material-ui/core/Card';
import CardHeader from '@material-ui/core/CardHeader';

import { AlertDialog, PopoverDialog } from './comp.jsx';
import { DragNdrop } from './dragzone.jsx';
import { KeyEdit, PwdDialog } from './compUI.jsx';

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
	const { children, useStyles, stopParamFn, pullFn, masterKeys, setMasterKeys, ...other } = props;
	const classes = useStyles();
	const [anchorEl, setAnchorEl] = useState(null);

	// popover for stop
	const handleClick = (ev, val) => {
		setAnchorEl({
			el: ev.currentTarget,
			val: val,
		});
	};
	const handleStop = () => {
		console.log('[rm]', anchorEl.val);
		const val = anchorEl.val;
		const param = stopParamFn(val);
		fetchReq(param.url, param.param, (d) => {
			console.log('[rm][stop]', d);
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
				// console.log('[pull][key]', d);
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
			<PopoverDialog
				data={anchorEl}
				setData={setAnchorEl}
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
	const [isAddMode, setAddMode] = useState(false);

	const fileRef = useRef();
	const [masterKeys, setMasterKeys] = useState([]);
	const [editData, setEditData] = useState({});

	const [dialogData, setDialog] = useState(null);
	const [pwdDialog, setPwdDialog] = useState(null);
	const [popover, setPopover] = useState(null);

	// add req & cancel
	const handleCancel = () => {
		setEditData({});
		setAddMode(false);
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
			uuid: v?.node?.split('/')[0],
			key: v?.key,
		};
		console.log('[key][add]', param.uuid);
		fetchReq('./api/key/?op=set', {
			body: JSON.stringify(param),
			method: 'POST',
		}, (d) => {
			console.log('[key][add]ret', d);
			handleCancel();
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

			// Dialog for password
			if (cryptoApi.isEncrypt(json)) {
				setPwdDialog({
					title: '檔案已加密',
					cb: (ev, val) => handleFileDec(json, val),
				});
				return;
			}

			loadKeys(json);
		}
		reader.readAsText(val[0]);
	};
	const handleFileDec = async (json, pwd) => {
		// TODO: decrypt in client? or send encrypted data
		try { // decrypt data
			const obj = await cryptoApi.decrypt(json, pwd);
			loadKeys(obj);
		} catch (err) {
			setDialog({
				title: 'Error',
				msg: err.toString(),
			});
		}
	}
	const loadKeys = (json) => {
		let reqs = [];
		json?.keys?.forEach((v, i) => {
			let param = {
				uuid: v.tag,
				key: v.key,
			};
			reqs.push(fetch('./api/key/?op=set', {
				body: JSON.stringify(param),
				method: 'POST',
			}));
		});

		reqs.length && Promise.allSettled(reqs).then((rets) => {
			// {status: "fulfilled", value: [...]}
			// console.log('[load]key state', rets);
			// TODO: error handle
			reqs = rets.filter((res) => res.status === 'fulfilled').map((res) => res.value.json());
			reqs.length && Promise.allSettled(reqs).then((rets) => {
				console.log('[load]key state: json', rets);
				rets = rets.filter((res) => res.status === 'fulfilled').map((res) => res.value);
				let last = rets[0];
				rets.forEach(ret => {
					if (ret.length > last.length) last = ret;
				});
				if (last.sort) last.sort();
				setMasterKeys(last);
			});
		});
	}

	const handleClearBtn = (e) => {
		setPopover({
			el: e.currentTarget,
		});
	}
	const handleClear = (e) => {
		fetchReq('./api/key/?op=clr', {
			method: 'POST',
		}, (d) => {
			setMasterKeys(d);
		}, (err) => {
			setDialog({
				title: 'Error',
				msg: err,
			});
		});
	}

	return (
		<div>
			<AlertDialog data={dialogData} setDialog={setDialog}></AlertDialog>
			<PwdDialog data={pwdDialog} setDialog={setPwdDialog} />
			{!isAddMode &&
				<DragNdrop ref={fileRef} handleFile={handleFile} onClick={false}>
					<Tooltip title="Add" aria-label="add">
						<Fab color="primary" className={classes.addBtn} onClick={() => setAddMode(true)}>
							<AddIcon />
						</Fab>
					</Tooltip>
					<Tooltip title="Load" aria-label="load">
						<Fab color="primary" className={classes.addBtn} onClick={handleLoadBtn}>
							<FolderOpenIcon />
						</Fab>
					</Tooltip>
					<Tooltip title="Clear all" aria-label="clear all">
						<Fab color="secondary" className={classes.addBtn} onClick={handleClearBtn}>
							<DeleteSweepIcon />
						</Fab>
					</Tooltip>
					<PanelListKeys
						useStyles={useStyles}
						stopParamFn={stopParamFn}
						pullFn={pullFn}
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
				<KeyEdit
					isNew={true}
					showNote={false}
					editData={editData}
					setEditData={setEditData}
					onCancel={handleCancel}
					onAdd={handleAdd}
				/>
			}
		</div>
	);
}

export { KeyPanel };
