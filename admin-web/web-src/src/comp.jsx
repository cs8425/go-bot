import { h, Fragment } from 'preact';
import { useState, useEffect, useContext } from 'preact/hooks';

import { makeStyles } from '@material-ui/core/styles';

import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableContainer from '@material-ui/core/TableContainer';
import TableHead from '@material-ui/core/TableHead';
import TablePagination from '@material-ui/core/TablePagination';
import TableRow from '@material-ui/core/TableRow';

function DataList(props) {
	const { children, data, header, renderRow, ...other } = props;
	const [page, setPage] = useState(0);
	const [rowsPerPage, setRowsPerPage] = useState(10);

	const handleChangePage = (event, newPage) => {
		setPage(newPage);
	};

	const handleChangeRowsPerPage = (event) => {
		setRowsPerPage(+event.target.value);
		setPage(0);
	};

	let rows = data || [];
	return (
		<>
			<TableContainer>
				<Table stickyHeader aria-label="sticky table">
					<TableHead>{header}</TableHead>
					<TableBody>
						{rows.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage).map((row, idx) => {
							return renderRow(row, idx);
						})}
					</TableBody>
				</Table>
			</TableContainer>
			<TablePagination
				rowsPerPageOptions={[10, 25, 50, 100]}
				component="div"
				count={rows.length}
				rowsPerPage={rowsPerPage}
				page={page}
				onChangePage={handleChangePage}
				onChangeRowsPerPage={handleChangeRowsPerPage}
			/>
		</>
	);
}

export { DataList };

import Button from '@material-ui/core/Button';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';

function AlertDialog(props) {
	const { children, data, setDialog, ...other } = props;
	const handleClose = () => {
		setDialog(null);
	};

	return (
		<div>
			<Dialog
				open={data !== null}
				onClose={handleClose}
				aria-labelledby="alert-dialog-title"
				aria-describedby="alert-dialog-description"
			>
				{data?.title &&
					<DialogTitle id="alert-dialog-title">{data?.title}</DialogTitle>
				}
				{data?.msg &&
					<DialogContent>
						<DialogContentText id="alert-dialog-description">{children || data?.msg}</DialogContentText>
					</DialogContent>
				}
				<DialogActions>
					<Button onClick={handleClose} color="primary" autoFocus>OK</Button>
				</DialogActions>
			</Dialog>
		</div>
	);
}

export { AlertDialog };

import Tooltip from '@material-ui/core/Tooltip';
import Popover from '@material-ui/core/Popover';

import Box from '@material-ui/core/Box';
import Fab from '@material-ui/core/Fab';
import ButtonGroup from '@material-ui/core/ButtonGroup';

const useStyles = makeStyles((theme) => ({
	popover: {
		margin: theme.spacing(2),
	},
	noUppercase: {
		textTransform: 'unset',
	},
}));

function PopoverDialog(props) {
	const { children, data, setData, onConfirm, onClose, ...other } = props;
	const classes = useStyles();
	const handleClose = () => {
		if (typeof onClose === 'function') {
			let ret = onClose(e);
			if (ret != false) setData(null);
			return;
		}
		setData(null);
	};
	const handleConfirm = (e) => {
		if (typeof onConfirm === 'function') {
			let ret = onConfirm(e);
			if (ret != false) setData(null);
			return;
		}
		setData(null);
	}

	return (
		<Popover
			open={data !== null}
			onClose={handleClose}
			anchorEl={data?.el}
			anchorOrigin={{
				vertical: 'top',
				horizontal: 'left',
			}}
			transformOrigin={{
				vertical: 'top',
				horizontal: 'left',
			}}
			{...other}
		>
			<Box className={classes.popover}>
				{children}
				<ButtonGroup disableElevation variant="contained">
					<Button className={classes.noUppercase} onClick={handleClose}>Cancel</Button>
					<Button className={classes.noUppercase} onClick={handleConfirm} color="secondary" >Remove</Button>
				</ButtonGroup>
			</Box>
		</Popover>
	);
}

export { PopoverDialog };


import AddIcon from '@material-ui/icons/Add';

import { fetchReq } from './api.js';

function PanelListMode(props) {
	const { children, useStyles, handleAddBtn, stopParamFn, ksParamFn, pullFn, dataStore, header, renderRowFn, ...other } = props;
	const classes = useStyles();
	const [popover, setPopover] = useState(null);
	const srvStore = useContext(dataStore);

	// popover for stop
	const handleClick = (ev, val) => {
		console.log('[popover]', ev, val);
		setPopover({
			el: ev.currentTarget,
			val: val,
		});
	};
	const handleStop = () => {
		console.log('[stop]', popover.val);
		const val = popover.val;
		const param = stopParamFn(val);
		fetchReq(param.url, param.param, (d) => {
			console.log('[rev][stop]', d);
			srvStore.set(d);
		}, (err) => {
			console.log('[rev][stop]err', err);
			setDialog({
				title: 'Error',
				msg: err,
			});
		});
	}
	const handleKS = (e, val) => {
		console.log('[KS]', e, val);
		const param = ksParamFn(val);
		fetchReq(param.url, param.param, (d) => {
			console.log('[rev][ks]', d);
			srvStore.set(d);
		}, (err) => {
			console.log('[rev][ks]err', err);
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
			// console.log('[pull][rev]', intv);
			pullFn()?.then((d) => {
				// console.log(d);
				srvStore.set(d);
			});
			t = setTimeout(pull, intv);
		};
		pull();
		return () => {
			clearTimeout(t);
			// console.log('[pull][rev]stop');
		};
	}, [props.interval]);

	const renderRow = (v, idx) => {
		return renderRowFn({
			v,
			idx,
			onClick: (e) => handleClick(e, v),
			onKillSwitch: (e) => handleKS(e, v),
		});
	}

	return (
		<div>
			<PopoverDialog
				data={popover}
				setData={setPopover}
				// onClose={handleClose}
				onConfirm={handleStop}
			>
				<p>確定要停止嗎?</p>
			</PopoverDialog>

			<Tooltip title="Add" aria-label="add">
				<Fab color="primary" className={classes.addBtn} onClick={handleAddBtn}>
					<AddIcon />
				</Fab>
			</Tooltip>
			<DataList header={header} renderRow={renderRow} data={srvStore.val}></DataList>
		</div>
	);
}

export { PanelListMode };
