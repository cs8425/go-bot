import { h, Fragment, Component } from 'preact';
import { useState } from 'preact/hooks';

import { NodeStore } from './store.js';

// 引入組件
import { Counter } from './comp.jsx';
import { LocalPanel } from './compLocal.jsx';
import { NodePanel } from './compNode.jsx';

import AppBar from '@material-ui/core/AppBar';
import Tab0 from '@material-ui/core/Tab';
import Tabs from '@material-ui/core/Tabs';
import TabPanel from './Tabs.jsx';

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
	const [nodeStore, setNodeStore] = useState(1);

	const handleTabChange = (event, newValue) => {
		setCurrTab(newValue);
	};

	return (
		<div className={classes.root}>
			<AppBar position="static">
				<Tabs
					value={currTab}
					onChange={handleTabChange}
					aria-label="tabs"
				>
					<Tab label="Nodes" {...a11yProps(0)} />
					<Tab label="Local bind" {...a11yProps(1)} />
					<Tab label="Remote bind" {...a11yProps(2)} />
				</Tabs>
			</AppBar>

			<NodeStore.Provider value={nodeStore}>
				<TabPanel value={currTab} index={0}>
					<NodePanel setNodeStore={setNodeStore}></NodePanel>
				</TabPanel>
				<TabPanel value={currTab} index={1}>
					<LocalPanel NodeStore={NodeStore}></LocalPanel>
				</TabPanel>
				<TabPanel value={currTab} index={2}>Item Three<Counter></Counter>
					<LocalPanel NodeStore={NodeStore}></LocalPanel>
				</TabPanel>
			</NodeStore.Provider>
		</div>
	);
}

export default app;
export { app as App };
