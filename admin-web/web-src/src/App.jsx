import { h, Fragment, Component } from 'preact';
import { useState, useRef } from 'preact/hooks';

import { NodeStore, LocalStore, RevStore } from './store.js';

// 引入組件
import { NodePanel } from './compNode.jsx';
import { LocalPanel } from './compLocal.jsx';
import { ReversePanel } from './compReverse.jsx';
import { DragNdrop } from './dragzone.jsx';

import AppBar from '@material-ui/core/AppBar';
import Tab0 from '@material-ui/core/Tab';
import Tabs from '@material-ui/core/Tabs';
import TabPanel from './Tabs.jsx';

import SaveIcon from '@material-ui/icons/Save';
import FolderOpenIcon from '@material-ui/icons/FolderOpen';
import Tooltip from '@material-ui/core/Tooltip';

import { makeStyles, styled } from "@material-ui/core/styles";

function a11yProps(index) {
	return {
		id: `tab-${index}`,
		"aria-controls": `tabpanel-${index}`,
	};
}

const styles = {
	root: {
		flexGrow: 1,
		// backgroundColor: theme.palette.background.paper,
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
	const [nodeStore, setNodeStore] = useState([]);
	const [localStore, setLocalStore] = useState([]);
	const [revStore, setRevStore] = useState([]);
	const dummyDlEl = useRef(null);
	const fileRef = useRef();

	const handleTabChange = (event, newValue) => {
		if (newValue > 2) {
			console.log('[load/save]', newValue);
			return;
		}
		setCurrTab(newValue);
	};
	const handleSave = (e) => {
		Promise.allSettled([
			fetch('./api/local/').then(function (res) {
				return res.json();
			}),
			fetch('./api/rev/').then(function (res) {
				return res.json();
			})
		]).then(([localRet, revRet]) => {
			// {status: "fulfilled", value: [...]}
			console.log('[save]state', localRet, revRet);
			if (localRet.status !== 'fulfilled' || revRet.status !== 'fulfilled') {
				return; // TODO: alert
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
			dumpJson(dump);
		});
	}
	const dumpJson = (data, fileName) => {
		// console.log('[dump]', data);
		const el = dummyDlEl.current;
		let json = JSON.stringify(data);
		let blob = new Blob([json], { type: "octet/stream" });
		let url = window.URL.createObjectURL(blob);
		el.href = url;
		el.download = fileName || 'config.json';
		el.click();
		setTimeout(() => { window.URL.revokeObjectURL(url); }, 30 * 1000);
	};
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
				let last = rets.pop()
				last.value.json().then(function (d) {
					setLocalStore(d);
				});
			});

			Promise.allSettled(revReqs).then((rets) => {
				// {status: "fulfilled", value: [...]}
				console.log('[load]rev state', rets);
				let last = rets.pop()
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

					<Tooltip title="Load" aria-label="load"><Tab onClick={handleLoadBtn} icon={<FolderOpenIcon />} {...a11yProps(3)} /></Tooltip>
					<Tooltip title="Save" aria-label="save"><Tab onClick={handleSave} icon={<SaveIcon />} {...a11yProps(4)} /></Tooltip>
				</Tabs>
			</AppBar>

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

			{/* dummy link for download file */}
			<a style="display: none;" ref={dummyDlEl}></a>
		</div>
	);
}

export default app;
export { app as App };
