import { h, Fragment, Component } from 'preact';
import { useState, useRef } from 'preact/hooks';

import { NodeStore, LocalStore, RevStore, KeyStore } from './store.js';
import { dumpJson } from './api.js';

// 引入組件
import { NodePanel } from './compNode.jsx';
import { LocalPanel } from './compLocal.jsx';
import { ReversePanel } from './compReverse.jsx';
import { KeyPanel } from './compKey.jsx';
import { KeyEditPanel } from './compKeyEdit.jsx';
import { DragNdrop } from './dragzone.jsx';

import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Tab0 from '@material-ui/core/Tab';
import Tabs from '@material-ui/core/Tabs';
import TabPanel from './Tabs.jsx';
import Tooltip from '@material-ui/core/Tooltip';
import Switch from '@material-ui/core/Switch';
import Grid from '@material-ui/core/Grid';

import SaveIcon from '@material-ui/icons/Save';
import FolderOpenIcon from '@material-ui/icons/FolderOpen';
import VpnKeyIcon from '@material-ui/icons/VpnKey';

import { makeStyles, styled } from "@material-ui/core/styles";

function a11yProps(index) {
	return {
		id: `tab-${index}`,
		"aria-controls": `tabpanel-${index}`,
		'value': index,
	};
}

const styles = {
	root: {
		flexGrow: 1,
		// backgroundColor: theme.palette.background.paper,
	},
	grow: {
		flexGrow: 1,
	},
	tab: {
		textTransform: 'unset',
	},
};
const useStyles = makeStyles((theme) => (styles));
const Tab = styled(Tab0)(styles.tab);

function app() {
	const classes = useStyles();
	const [currTab, setCurrTab] = useState(0);
	const [currMode, setCurrMode] = useState(false);
	const [nodeStore, setNodeStore] = useState([]);
	const [localStore, setLocalStore] = useState([]);
	const [revStore, setRevStore] = useState([]);
	const [keyStore, setKeyStore] = useState([]);
	const dummyDlEl = useRef(null);
	const fileRef = useRef();

	const handleTabChange = (event, newValue) => {
		switch (newValue) {
			case 'load':
			case 'save':
				console.log('[load/save]', newValue);
				return;
		}
		setCurrTab(newValue);
	};
	const handleSave = (e) => {
		Promise.allSettled([
			fetch('./api/local/').then((res) => res.json()),
			fetch('./api/rev/').then((res) => res.json()),
		]).then(([localRet, revRet]) => {
			// {status: "fulfilled", value: [...]}
			console.log('[save]state', localRet, revRet);
			if (localRet.status !== 'fulfilled' || revRet.status !== 'fulfilled') {
				return; // TODO: error handle
			}
			const dump = {
				local: localRet.value?.map((v, i) => ({
					addr: v.addr,
					args: v.args,
					tag: v.id.split('/')[0],
				})),
				rev: revRet.value?.map((v, i) => ({
					addr: v.addr,
					target: v.target,
					args: v?.args,
					tag: v.id.split('/')[0],
				})),
			};
			dumpJson(dummyDlEl.current, dump);
		});
	}
	const handleLoadBtn = (e) => {
		console.log('[load]click', e);
		fileRef.current.open();
	}
	const getNode = (tag) => {
		return nodeStore?.find(v => v.tag.match(tag))?.tag;
	}
	const handleFile = (val) => {
		console.log('[file]', val);
		let reader = new FileReader();
		reader.onload = (e) => {
			const json = JSON.parse(e.target.result);
			console.log(e, json);
			let loReqs = [];
			json?.local?.forEach((v, i) => {
				const useNode = getNode(v.tag);
				if (!useNode) return;
				let param = {
					uuid: useNode,
					bind_addr: v.addr,
					argv: v.args,
					pause: true,
				};
				loReqs.push(fetch('./api/local/?op=bind', {
					body: JSON.stringify(param),
					method: 'POST',
				}));
			});

			let revReqs = [];
			json?.rev?.forEach((v, i) => {
				const useNode = getNode(v.tag);
				if (!useNode) return;
				let param = {
					uuid: useNode,
					remote: v.addr,
					target: v.target,
					argv: v.args,
					pause: true,
				};
				revReqs.push(fetch('./api/rev/?op=bind', {
					body: JSON.stringify(param),
					method: 'POST',
				}));
			});

			Promise.allSettled(loReqs).then((rets) => {
				// {status: "fulfilled", value: [...]}
				console.log('[load]local state', rets);
				let last = rets.pop();
				// TODO: error handle
				last.value.json().then(function (d) {
					setLocalStore(d);
				});
			});

			Promise.allSettled(revReqs).then((rets) => {
				// {status: "fulfilled", value: [...]}
				console.log('[load]rev state', rets);
				let last = rets.pop();
				// TODO: error handle
				last.value.json().then(function (d) {
					setRevStore(d);
				});
			});
		}
		reader.readAsText(val[0]);
	};

	return (
		<div className={classes.root}>
			<AppBar position="static">
				<Toolbar variant="dense">
					{currMode == false &&
						<Tabs
							value={currTab}
							onChange={handleTabChange}
							variant="scrollable"
							scrollButtons="auto"
							aria-label="tabs"
						>
							<Tab label="Nodes" {...a11yProps(0)} />
							<Tab label="Local bind" {...a11yProps(1)} />
							<Tab label="Remote bind" {...a11yProps(2)} />

							<Tooltip title="Import Keys" aria-label="import keys" {...a11yProps('key')}><Tab icon={<VpnKeyIcon />} /></Tooltip>

							<Tooltip title="Load" aria-label="load" {...a11yProps('load')}><Tab onClick={handleLoadBtn} icon={<FolderOpenIcon />} /></Tooltip>
							<Tooltip title="Save" aria-label="save" {...a11yProps('save')}><Tab onClick={handleSave} icon={<SaveIcon />} /></Tooltip>
						</Tabs>
					}
					{currMode == true &&
						<Tabs
							value={currTab}
							onChange={handleTabChange}
							variant="scrollable"
							scrollButtons="auto"
							aria-label="tabs"
						>
							<Tooltip title="Keys" aria-label="keys" {...a11yProps('key')}><Tab icon={<VpnKeyIcon />} /></Tooltip>

						</Tabs>
					}

					<div className={classes.grow} /> {/* space */}

					<div>
						<Grid component="div" container alignItems="center" spacing={1}>
							<Grid item>使用</Grid>
							<Grid item>
								<Switch checked={currMode} onChange={() => { setCurrMode(!currMode); setCurrTab(!currMode ? 'key': 0) }} name="editorSwitch" />
							</Grid>
							<Grid item>編輯</Grid>
						</Grid>
					</div>

				</Toolbar>
			</AppBar>

			{currMode == false &&
				<>
					<DragNdrop ref={fileRef} handleFile={handleFile} onClick={false}>
						<NodeStore.Provider value={nodeStore}>
							<TabPanel value={currTab} index={0}>
								<NodePanel setNodeStore={setNodeStore}></NodePanel>
							</TabPanel>
							<TabPanel value={currTab} index={1}>
								<LocalStore.Provider value={{ val: localStore, set: setLocalStore }}><LocalPanel /></LocalStore.Provider>
							</TabPanel>
							<TabPanel value={currTab} index={2}>
								<RevStore.Provider value={{ val: revStore, set: setRevStore }}><ReversePanel /></RevStore.Provider>
							</TabPanel>
						</NodeStore.Provider>
					</DragNdrop>
					<NodeStore.Provider value={nodeStore}>
						<TabPanel value={currTab} index={'key'}>
							<KeyStore.Provider value={{ val: keyStore, set: setKeyStore }}><KeyPanel /></KeyStore.Provider>
						</TabPanel>
					</NodeStore.Provider>
				</>
			}
			{currMode == true &&
				// local editor mode, be careful with sensitivity value
				<NodeStore.Provider value={nodeStore}>
					<TabPanel value={currTab} index={'key'}>
						<KeyEditPanel />
					</TabPanel>
				</NodeStore.Provider>
			}

			{/* dummy link for download file */}
			<a style="display: none;" ref={dummyDlEl}></a>
		</div>
	);
}

export default app;
export { app as App };
