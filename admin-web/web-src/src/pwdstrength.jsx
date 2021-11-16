import { h, Fragment, Component } from 'preact';
import { useState } from 'preact/hooks';

import { withStyles } from '@material-ui/core/styles';
import LinearProgress from '@material-ui/core/LinearProgress';
import Typography from '@material-ui/core/Typography';
import Box from '@material-ui/core/Box';

const ColorLinearProgress = withStyles((theme) => ({
	root: {
		height: 10,
		borderRadius: 5,
	},
	colorPrimary: {
		backgroundColor: theme.palette.grey[theme.palette.type === 'light' ? 200 : 700],
	},
	bar: {
		borderRadius: 5,
		backgroundColor: (props) => props.color,
	},
}))(LinearProgress);

const strengthMap = {
	0: {
		text: 'Danger',
		color: '#f11a1a',
		val: 5,
	},
	1: {
		text: 'Weak',
		color: '#ffdd57',
		val: 33,
	},
	2: {
		text: 'Medium',
		color: '#1a90ff',
		val: 66,
	},
	3: {
		text: 'Strong',
		color: '#48c774',
		val: 100,
	},
};

function LinearProgressWithLabel(props) {
	const { value, style, ...other } = props;

	const hasType = [
		/[a-z]/,
		/[A-Z]/,
		/[0-9]/,
		/[\!@\#\$%\^&\*`~\-_\=\+\]\[\{\}'"\/\\\?\.\>,\<\(\)]/,
	].filter((regex) => regex.test(value));
	let lv = strengthMap[0];
	if (value.length >= 4 && hasType.length >= 2) lv = strengthMap[1];
	if (value.length >= 10 && hasType.length >= 3) lv = strengthMap[2];
	if (value.length >= 14 && hasType.length >= 4) lv = strengthMap[3];
	// console.log('[pwd]', hasType, hasType.length, lv);

	return (
		<Box display="flex" alignItems="center" {...style}>
			<Box width="100%" mr={1}>
				<ColorLinearProgress color={lv.color} variant="determinate" value={lv.val} {...other} />
			</Box>
			<Box minWidth={35}>
				<Typography variant="body2" color="textSecondary">{`${lv.text}`}</Typography>
			</Box>
		</Box>
	);
}

export { LinearProgressWithLabel };
