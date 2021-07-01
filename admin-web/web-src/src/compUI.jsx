import { h, Fragment } from 'preact';
import { useState, useContext } from 'preact/hooks';

import { NodeStore } from './store.js';

import { makeStyles } from '@material-ui/core/styles';

import Box from '@material-ui/core/Box';
import Button from '@material-ui/core/Button';
import ButtonGroup from '@material-ui/core/ButtonGroup';

import TextField from '@material-ui/core/TextField';
import MenuItem from '@material-ui/core/MenuItem';
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
