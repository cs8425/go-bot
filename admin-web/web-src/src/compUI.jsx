import { h, Fragment } from 'preact';
import { useState, useContext } from 'preact/hooks';

import { NodeStore } from './store.js';

import { makeStyles } from '@material-ui/core/styles';

import Box from '@material-ui/core/Box';
import Button from '@material-ui/core/Button';
import ButtonGroup from '@material-ui/core/ButtonGroup';

import TextField from '@material-ui/core/TextField';
import MenuItem from '@material-ui/core/MenuItem';
import InputAdornment from '@material-ui/core/InputAdornment';
import Typography from '@material-ui/core/Typography';

import AddCircleIcon from '@material-ui/icons/AddCircle';
import CancelIcon from '@material-ui/icons/Cancel';
import DeleteIcon from '@material-ui/icons/Delete';
import SaveIcon from '@material-ui/icons/Save';

import { AlertDialog } from './comp.jsx';

const useStyles = makeStyles((theme) => ({
	center: {
		textAlign: 'center',
	},
	noUppercase: {
		textTransform: 'unset',
	},
	text: {
		color: theme.palette.text.primary,
	},
}));

function KeyEdit(props) {
	const { children, editData, setEditData, isNew, onCancel, onAdd, onSave, onRemove, showNote, ...other } = props;
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

			{(showNote !== false) &&
				<div style="margin: 1rem;">
					<TextField
						multiline
						label="Note"
						value={editData.note}
						onChange={(e) => setEditData({ ...editData, note: e.target.value })}
						helperText="註解"
					/>
				</div>
			}
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

export { KeyEdit };

import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';
import IconButton from '@material-ui/core/IconButton';

import Visibility from '@material-ui/icons/Visibility';
import VisibilityOff from '@material-ui/icons/VisibilityOff';

import { LinearProgressWithLabel } from './pwdstrength.jsx';

function PwdDialog(props) {
	const { children, data, setDialog, setPwd, ...other } = props;
	const initState = () => ({
		showPwd0: false,
		pwd0: '',

		showPwd1: false,
		pwd1: '',

		pwdMatch: true,

		hasError: false,
	});
	const [values, setValues] = useState(initState());

	const handleChange = (prop) => (e) => {
		// console.log('[val]', prop, e, e.target.value);
		setValues({ ...values, pwdMatch: true, [prop]: e.target.value });
	};
	const handleClickShowPassword = (prop) => () => {
		setValues({ ...values, [prop]: !values[prop] });
	};

	const handleClose = (e) => {
		setValues(initState());
		setDialog(null);
	};
	const handleOk = (e) => {
		if (setPwd) {
			if (values.pwd0 !== values.pwd1) {
				setValues({ ...values, pwdMatch: false });
				return;
			}
		}
		if (typeof data?.cb == 'function') {
			let ret = data.cb(e, values.pwd0);
			if (ret === false) return;
		}
		handleClose(e);
	};

	return (
		<Dialog
			open={data !== null}
			onClose={handleClose}
			maxWidth='lg'
		>
			{data?.title &&
				<DialogTitle>{data?.title}</DialogTitle>
			}

			<DialogContent dividers>
				{data?.msg &&
					<DialogContentText>{data?.msg}</DialogContentText>
				}

				<TextField
					variant="outlined"
					required
					autoFocus
					margin="dense"
					label="Password"
					fullWidth
					InputProps={{
						endAdornment: (
							<InputAdornment position="end">
								<IconButton
									aria-label="toggle password visibility"
									onClick={handleClickShowPassword('showPwd0')}
									onMouseDown={(e) => e.preventDefault()}
									edge="end"
								>
									{values.showPwd0 ? <Visibility /> : <VisibilityOff />}
								</IconButton>
							</InputAdornment>
						),
					}}
					type={values.showPwd0 ? 'text' : 'password'}
					value={values.pwd0}
					onChange={handleChange('pwd0')}
					error={!values.pwdMatch}
				/>

				{setPwd &&
					<>
						{!!values.pwd0 &&
							<LinearProgressWithLabel style={{ 'display': 'inline-flex', 'width': '100%' }} value={values.pwd0} />
						}

						<TextField
							variant="outlined"
							required
							margin="dense"
							label="Password Again"
							fullWidth
							InputProps={{
								endAdornment: (
									<InputAdornment position="end">
										<IconButton
											aria-label="toggle password visibility"
											onClick={handleClickShowPassword('showPwd1')}
											onMouseDown={(e) => e.preventDefault()}
											edge="end"
										>
											{values.showPwd1 ? <Visibility /> : <VisibilityOff />}
										</IconButton>
									</InputAdornment>
								),
							}}
							type={values.showPwd1 ? 'text' : 'password'}
							value={values.pwd1}
							onChange={handleChange('pwd1')}
							helperText={(values.pwdMatch) ? null : 'Password not same!!'}
							error={!values.pwdMatch}
						/>
					</>
				}
			</DialogContent>

			<DialogActions>
				<Button onClick={handleClose} color="primary">Cancel</Button>
				<Button onClick={handleOk} color="primary">OK</Button>
			</DialogActions>
		</Dialog>
	);
}

export { PwdDialog };
